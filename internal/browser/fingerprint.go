package browser

import (
	"fmt"
	"embroidery-designs/internal/fingerprint"
)

func GenerateStealthScript(profile *fingerprint.BrowserProfile) string {
	return fmt.Sprintf(`
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined
		});
		
		Object.defineProperty(navigator, 'plugins', {
			get: () => [1, 2, 3, 4, 5]
		});
		
		Object.defineProperty(navigator, 'languages', {
			get: () => ['%s']
		});
		
		Object.defineProperty(screen, 'width', {
			get: () => %d
		});
		
		Object.defineProperty(screen, 'height', {
			get: () => %d
		});
		
		Object.defineProperty(screen, 'colorDepth', {
			get: () => %d
		});
		
		window.chrome = {
			runtime: {}
		};
		
		Object.defineProperty(navigator, 'permissions', {
			get: () => ({
				query: () => Promise.resolve({ state: 'granted' })
			})
		});
	`, 
		profile.Language,
		profile.ScreenWidth,
		profile.ScreenHeight,
		profile.ColorDepth,
	)
}

