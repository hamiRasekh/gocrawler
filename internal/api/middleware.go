package api

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"embroidery-designs/internal/auth"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		c.Next()
		
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		
		if raw != "" {
			path = path + "?" + raw
		}
		
		utils.GetLogger().Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.String("ip", clientIP),
			zap.Duration("latency", latency),
		)
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		utils.GetLogger().Error("Panic recovered",
			zap.Any("error", recovered),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(500, gin.H{
			"error": "Internal server error",
		})
		c.Abort()
	})
}

func CORSMiddleware(corsOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS origin from configuration
		origin := corsOrigin
		if origin == "" {
			origin = "*"
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		
		// Only allow credentials if origin is not wildcard
		if origin != "*" {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

func JWTAuthMiddleware(cfg *config.Config, repository *storage.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate JWT token
		claims, err := auth.ValidateToken(cfg, tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if this is an API token by hashing and looking it up
		tokenHash := hashToken(tokenString)
		apiToken, err := repository.GetAPITokenByHash(c.Request.Context(), tokenHash)
		if err == nil {
			// This is an API token - update last used time
			_ = repository.UpdateTokenLastUsed(c.Request.Context(), apiToken.ID)
		}
		// If not found in DB, it's a login token (which is fine)

	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Next()
	}
}

// IP-based rate limiter for authentication endpoints
type ipRateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

func newIPRateLimiter(rate int, window time.Duration) *ipRateLimiter {
	rl := &ipRateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	
	// Cleanup old entries periodically
	go rl.cleanup()
	
	return rl
}

func (rl *ipRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *ipRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	v, exists := rl.visitors[ip]
	
	if !exists {
		rl.visitors[ip] = &visitor{
			count:    1,
			lastSeen: now,
		}
		return true
	}
	
	// Reset if window has passed
	if now.Sub(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = now
		return true
	}
	
	// Check if limit exceeded
	if v.count >= rl.rate {
		return false
	}
	
	v.count++
	v.lastSeen = now
	return true
}

// AuthRateLimitMiddleware creates a rate limiting middleware for authentication endpoints
func AuthRateLimitMiddleware(cfg *config.Config) gin.HandlerFunc {
	rate := cfg.Auth.RateLimitRequests
	window := cfg.Auth.RateLimitWindow
	
	// Default values if not configured
	if rate <= 0 {
		rate = 5
	}
	if window <= 0 {
		window = 15 * time.Minute
	}
	
	limiter := newIPRateLimiter(rate, window)
	
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		if !limiter.allow(ip) {
			utils.GetLogger().Warn("Rate limit exceeded for auth endpoint",
				zap.String("ip", ip),
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

