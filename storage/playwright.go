package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
)

func DownloadPhotosWithPlaywright(urlsString, outputDir string, baseIndex int) int {
	if strings.TrimSpace(urlsString) == "" {
		return 0
	}

	pw, err := playwright.Run()
	if err != nil {
		fmt.Printf("[DEBUG] Playwright run error: %v\n", err)
		return 0
	}
	defer func() { _ = pw.Stop() }()

	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}

	userDataDir := filepath.Join(localAppData, "Google", "Chrome", "User Data")

	var context playwright.BrowserContext

	if _, err := os.Stat(userDataDir); err == nil {
		context, err = pw.Chromium.LaunchPersistentContext(userDataDir, playwright.BrowserTypeLaunchPersistentContextOptions{
			Headless: playwright.Bool(true),
			Args: []string{
				"--no-sandbox",
				"--disable-setuid-sandbox",
				"--disable-dev-shm-usage",
			},
		})
		if err != nil {
			fmt.Printf("[DEBUG] Playwright persistent context error: %v\n", err)
			context = nil
		}
	} else {
		fmt.Printf("[DEBUG] Chrome user data dir not found at %s: %v\n", userDataDir, err)
	}

	// Fallback: launch regular browser without persistent context
	if context == nil {
		fmt.Printf("[DEBUG] Falling back to non-persistent browser launch\n")
		browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(true),
			Args: []string{
				"--no-sandbox",
				"--disable-setuid-sandbox",
				"--disable-dev-shm-usage",
			},
		})
		if err != nil {
			fmt.Printf("[DEBUG] Playwright browser launch error: %v\n", err)
			return 0
		}
		defer func() { _ = browser.Close() }()

		context, err = browser.NewContext()
		if err != nil {
			fmt.Printf("[DEBUG] Playwright context error: %v\n", err)
			return 0
		}
	}

	defer func() { _ = context.Close() }()

	page, err := context.NewPage()
	if err != nil {
		fmt.Printf("[DEBUG] Playwright page error: %v\n", err)
		return 0
	}
	defer func() { _ = page.Close() }()

	// Inject cookies if we have them
	if cookies, err := GetBrowserCookies("avito.ru"); err == nil && len(cookies) > 0 {
		fmt.Printf("[DEBUG] Injecting %d cookies into Playwright context\n", len(cookies))
		playwrightCookies := make([]playwright.OptionalCookie, 0, len(cookies))
		for name, value := range cookies {
			playwrightCookies = append(playwrightCookies, playwright.OptionalCookie{
				Name:   name,
				Value:  value,
				Domain: playwright.String(".avito.ru"),
				Path:   playwright.String("/"),
			})
		}
		_ = context.AddCookies(playwrightCookies)
	}

	urls := strings.Split(urlsString, "|")
	downloaded := 0

	for i, rawURL := range urls {
		url := strings.TrimSpace(rawURL)
		if url == "" {
			continue
		}

		response, err := page.Goto(url, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			Timeout:   playwright.Float(30000),
		})
		if err != nil {
			fmt.Printf("[DEBUG] Playwright goto error base=%d idx=%d: %v\n", baseIndex, i, err)
			continue
		}

		time.Sleep(500 * time.Millisecond)

		if response == nil {
			fmt.Printf("[DEBUG] Playwright nil response base=%d idx=%d\n", baseIndex, i)
			continue
		}

		if !response.Ok() {
			fmt.Printf("[DEBUG] Playwright bad status base=%d idx=%d: status=%d\n", baseIndex, i, response.Status())
			continue
		}

		body, err := response.Body()
		if err != nil {
			fmt.Printf("[DEBUG] Playwright body error base=%d idx=%d: %v\n", baseIndex, i, err)
			continue
		}

		if len(body) < 100 {
			fmt.Printf("[DEBUG] Playwright response too small base=%d idx=%d: %d bytes\n", baseIndex, i, len(body))
			continue
		}

		ext := ".jpg"
		ct, _ := response.HeaderValue("content-type")
		if ct == "" {
			ctVal, err := page.Evaluate("document.contentType")
			if err == nil {
				if s, ok := ctVal.(string); ok {
					ct = s
				}
			}
		}
		if ct != "" {
			if strings.Contains(ct, "webp") {
				ext = ".webp"
			} else if strings.Contains(ct, "png") {
				ext = ".png"
			} else if strings.Contains(ct, "jpeg") || strings.Contains(ct, "jpg") {
				ext = ".jpg"
			}
		}

		savePath := filepath.Join(outputDir, fmt.Sprintf("a%d_%d%s", baseIndex, i, ext))
		if err := os.WriteFile(savePath, body, 0644); err != nil {
			fmt.Printf("[DEBUG] Playwright write error base=%d idx=%d: %v\n", baseIndex, i, err)
			continue
		}

		downloaded++
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("[DEBUG] Playwright downloaded base=%d count=%d\n", baseIndex, downloaded)
	return downloaded
}
