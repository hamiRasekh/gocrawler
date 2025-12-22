package fingerprint

import (
	"math/rand"
	"time"
)

type BrowserProfile struct {
	UserAgent      string
	AcceptLanguage string
	Accept         string
	AcceptEncoding string
	Connection     string
	UpgradeInsecureRequests string
	SecFetchSite   string
	SecFetchMode   string
	SecFetchUser   string
	SecFetchDest   string
	ViewportWidth  int
	ViewportHeight int
	ScreenWidth    int
	ScreenHeight   int
	ColorDepth     int
	Timezone       string
	Language       string
	Platform       string
}

var (
	userAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
	}
	
	acceptLanguages = []string{
		"en-US,en;q=0.9",
		"en-GB,en;q=0.9",
		"en-US,en;q=0.9,fa;q=0.8",
		"fa-IR,fa;q=0.9,en-US;q=0.8,en;q=0.7",
		"en-US,en;q=0.9,es;q=0.8",
	}
	
	timezones = []string{
		"America/New_York",
		"America/Los_Angeles",
		"Europe/London",
		"Europe/Paris",
		"Asia/Tehran",
		"Asia/Dubai",
		"Asia/Tokyo",
		"Australia/Sydney",
	}
	
	commonViewports = [][]int{
		{1920, 1080},
		{1366, 768},
		{1536, 864},
		{1440, 900},
		{1280, 720},
		{1600, 900},
	}
)

func GenerateProfile() *BrowserProfile {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	viewport := commonViewports[r.Intn(len(commonViewports))]
	
	return &BrowserProfile{
		UserAgent:      userAgents[r.Intn(len(userAgents))],
		AcceptLanguage: acceptLanguages[r.Intn(len(acceptLanguages))],
		Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		AcceptEncoding: "gzip, deflate, br",
		Connection:     "keep-alive",
		UpgradeInsecureRequests: "1",
		SecFetchSite:   "none",
		SecFetchMode:   "navigate",
		SecFetchUser:   "?1",
		SecFetchDest:   "document",
		ViewportWidth:  viewport[0],
		ViewportHeight: viewport[1],
		ScreenWidth:    viewport[0],
		ScreenHeight:   viewport[1],
		ColorDepth:     24,
		Timezone:       timezones[r.Intn(len(timezones))],
		Language:       "en-US",
		Platform:       "Win32",
	}
}

