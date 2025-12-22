package fingerprint

import (
	"net/http"
)

func ApplyHeaders(req *http.Request, profile *BrowserProfile) {
	req.Header.Set("User-Agent", profile.UserAgent)
	req.Header.Set("Accept-Language", profile.AcceptLanguage)
	req.Header.Set("Accept", profile.Accept)
	req.Header.Set("Accept-Encoding", profile.AcceptEncoding)
	req.Header.Set("Connection", profile.Connection)
	req.Header.Set("Upgrade-Insecure-Requests", profile.UpgradeInsecureRequests)
	req.Header.Set("Sec-Fetch-Site", profile.SecFetchSite)
	req.Header.Set("Sec-Fetch-Mode", profile.SecFetchMode)
	req.Header.Set("Sec-Fetch-User", profile.SecFetchUser)
	req.Header.Set("Sec-Fetch-Dest", profile.SecFetchDest)
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("DNT", "1")
}

func GetDefaultHeaders() map[string]string {
	profile := GenerateProfile()
	return map[string]string{
		"User-Agent":             profile.UserAgent,
		"Accept-Language":        profile.AcceptLanguage,
		"Accept":                 profile.Accept,
		"Accept-Encoding":        profile.AcceptEncoding,
		"Connection":             profile.Connection,
		"Upgrade-Insecure-Requests": profile.UpgradeInsecureRequests,
		"Sec-Fetch-Site":         profile.SecFetchSite,
		"Sec-Fetch-Mode":         profile.SecFetchMode,
		"Sec-Fetch-User":         profile.SecFetchUser,
		"Sec-Fetch-Dest":         profile.SecFetchDest,
		"Cache-Control":          "max-age=0",
		"DNT":                    "1",
	}
}

