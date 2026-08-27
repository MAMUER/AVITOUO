package storage

import (
	"fmt"
	"math"
	"math/rand"
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
			Headless: playwright.Bool(false),
			Channel:  playwright.String("chrome"),
			Args: []string{
				"--no-sandbox",
				"--disable-setuid-sandbox",
				"--disable-dev-shm-usage",
				"--start-maximized",
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
			Headless: playwright.Bool(false),
			Channel:  playwright.String("chrome"),
			Args: []string{
				"--no-sandbox",
				"--disable-setuid-sandbox",
				"--disable-dev-shm-usage",
				"--start-maximized",
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

	// Try to use browser cookies if available
	cookieCount := 0
	if cookies, err := GetBrowserCookies("avito.ru"); err == nil && len(cookies) > 0 {
		cookieCount = len(cookies)
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
	} else if cookies, err := GetBrowserCookies("www.avito.ru"); err == nil && len(cookies) > 0 {
		cookieCount = len(cookies)
		playwrightCookies := make([]playwright.OptionalCookie, 0, len(cookies))
		for name, value := range cookies {
			playwrightCookies = append(playwrightCookies, playwright.OptionalCookie{
				Name:   name,
				Value:  value,
				Domain: playwright.String("www.avito.ru"),
				Path:   playwright.String("/"),
			})
		}
		_ = context.AddCookies(playwrightCookies)
	}
	fmt.Printf("[DEBUG] Playwright context cookies injected=%d\n", cookieCount)

	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
	ua := userAgents[rand.Intn(len(userAgents))]
	_ = page.SetExtraHTTPHeaders(map[string]string{
		"User-Agent":      ua,
		"Accept-Language": "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
		"Referer":         "https://www.avito.ru/",
	})

	// Warm up session with avito.ru
	_, _ = page.Goto("https://www.avito.ru/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(45000),
	})
	time.Sleep(3 * time.Second)

	urls := strings.Split(urlsString, "|")
	downloaded := 0
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i, rawURL := range urls {
		url := strings.TrimSpace(rawURL)
		if url == "" {
			continue
		}

		response, err := page.Goto(url, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(45000),
		})
		if err != nil {
			fmt.Printf("[DEBUG] Playwright goto error base=%d idx=%d: %v\n", baseIndex, i, err)
			pause := time.Duration(2+rnd.Intn(3)) * time.Second
			time.Sleep(pause)
			continue
		}

		if response == nil {
			fmt.Printf("[DEBUG] Playwright nil response base=%d idx=%d\n", baseIndex, i)
			pause := time.Duration(2+rnd.Intn(3)) * time.Second
			time.Sleep(pause)
			continue
		}

		if !response.Ok() {
			fmt.Printf("[DEBUG] Playwright bad status base=%d idx=%d: status=%d\n", baseIndex, i, response.Status())
			if response.Status() == 429 {
				pause := time.Duration(5+rnd.Intn(5)) * time.Second
				fmt.Printf("[DEBUG] Playwright 429 backoff base=%d idx=%d: sleeping %v\n", baseIndex, i, pause)
				time.Sleep(pause)
			}
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
			pause := time.Duration(2+rnd.Intn(3)) * time.Second
			time.Sleep(pause)
			continue
		}

		downloaded++
		pause := time.Duration(2+rnd.Intn(4)) * time.Second
		time.Sleep(pause)
	}

	fmt.Printf("[DEBUG] Playwright downloaded base=%d count=%d\n", baseIndex, downloaded)
	return downloaded
}

func jitteredBackoff(attempt int, base time.Duration) time.Duration {
	if attempt <= 0 {
		return base
	}
	backoff := base * time.Duration(math.Pow(2, float64(attempt-1)))
	jitter := time.Duration(rand.Int63n(int64(base)))
	return backoff + jitter
}
