package browser

import (
	"context"

	"github.com/chromedp/chromedp"
	"embroidery-designs/internal/fingerprint"
)

func ApplyStealth(ctx context.Context, profile *fingerprint.BrowserProfile) error {
	script := GenerateStealthScript(profile)
	
	return chromedp.Run(ctx,
		chromedp.Evaluate(script, nil),
	)
}

