package websocket

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"embroidery-designs/internal/auth"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
)

func createUpgrader(cfg *config.Config) websocket.Upgrader {
	checkOrigin := func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		corsOrigin := cfg.Server.CORSOrigin
		
		// If CORS is set to wildcard or empty, allow all (development mode)
		if corsOrigin == "" || corsOrigin == "*" {
			return true
		}
		
		// Check if origin matches configured CORS origin
		return origin == corsOrigin
	}
	
	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024
)

func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request, taskID *int64, cfg *config.Config, repository *storage.Repository) {
	// Validate JWT token from query parameter or header
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}
	}
	
	if tokenString == "" {
		utils.GetLogger().Warn("WebSocket connection attempt without authentication")
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	
	// Validate token
	claims, err := auth.ValidateToken(cfg, tokenString)
	if err != nil {
		utils.GetLogger().Warn("WebSocket connection with invalid token", zap.Error(err))
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	
	// Check if this is an API token
	tokenHash := hashToken(tokenString)
	apiToken, err := repository.GetAPITokenByHash(r.Context(), tokenHash)
	if err == nil {
		// This is an API token - update last used time
		_ = repository.UpdateTokenLastUsed(r.Context(), apiToken.ID)
	}
	
	upgrader := createUpgrader(cfg)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.GetLogger().Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	client := &Client{
		ID:     generateClientID(),
		TaskID: taskID,
		UserID: claims.UserID,
		Username: claims.Username,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	client.Hub.register <- client

	go client.writePump(conn)
	go client.readPump(conn)
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (c *Client) readPump(conn *websocket.Conn) {
	defer func() {
		c.Hub.unregister <- c
		conn.Close()
	}()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetReadLimit(maxMessageSize)
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				utils.GetLogger().Error("WebSocket error", zap.Error(err))
			}
			break
		}
	}
}

func (c *Client) writePump(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func generateClientID() string {
	return time.Now().Format("20060102150405.000000000")
}

