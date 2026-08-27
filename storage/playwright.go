package storage

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mxschmitt/playwright-go"
	"golang.org/x/net/html"
)

var playwrightGlobalLock sync.Mutex

func DownloadPhotosWithPlaywright(urlsString, outputDir string, baseIndex int, proxy string) int {
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

	if context == nil {
		fmt.Printf("[DEBUG] Falling back to non-persistent browser launch\n")
		browserLaunchOpts := playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(false),
			Channel:  playwright.String("chrome"),
			Args: []string{
				"--no-sandbox",
				"--disable-setuid-sandbox",
				"--disable-dev-shm-usage",
				"--start-maximized",
			},
		}
		if strings.TrimSpace(proxy) != "" {
			browserLaunchOpts.Proxy = &playwright.Proxy{
				Server: strings.TrimSpace(proxy),
			}
			fmt.Printf("[DEBUG] Playwright using proxy: %s\n", proxy)
		}
		browser, err := pw.Chromium.Launch(browserLaunchOpts)
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

	_, _ = page.Goto("https://www.avito.ru/", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(45000),
	})
	time.Sleep(3 * time.Second)

	playwrightGlobalLock.Lock()
	defer playwrightGlobalLock.Unlock()

	urls := strings.Split(urlsString, "|")
	downloaded := 0
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	httpClient := &http.Client{
		Timeout: 45 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	for i, rawURL := range urls {
		url := strings.TrimSpace(rawURL)
		if url == "" {
			continue
		}

		response, err := page.Goto(url, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle,
			Timeout:   playwright.Float(60000),
		})
		if err != nil {
			fmt.Printf("[DEBUG] Playwright goto error base=%d idx=%d: %v\n", baseIndex, i, err)
			pause := time.Duration(20+rnd.Intn(20)) * time.Second
			time.Sleep(pause)
			continue
		}

		if response == nil {
			fmt.Printf("[DEBUG] Playwright nil response base=%d idx=%d\n", baseIndex, i)
			pause := time.Duration(20+rnd.Intn(20)) * time.Second
			time.Sleep(pause)
			continue
		}

		if !response.Ok() {
			fmt.Printf("[DEBUG] Playwright bad status base=%d idx=%d: status=%d\n", baseIndex, i, response.Status())
			if response.Status() == 429 {
				pause := time.Duration(90+rnd.Intn(60)) * time.Second
				fmt.Printf("[DEBUG] Playwright 429 backoff base=%d idx=%d: sleeping %v\n", baseIndex, i, pause)
				time.Sleep(pause)
			} else {
				pause := time.Duration(20+rnd.Intn(20)) * time.Second
				time.Sleep(pause)
			}
			continue
		}

		imageURLs, err := extractImageURLs(page)
		if err != nil {
			fmt.Printf("[DEBUG] Playwright extract images error base=%d idx=%d: %v\n", baseIndex, i, err)
			pause := time.Duration(20+rnd.Intn(20)) * time.Second
			time.Sleep(pause)
			continue
		}

		if len(imageURLs) == 0 {
			fmt.Printf("[DEBUG] Playwright no images found base=%d idx=%d url=%s\n", baseIndex, i, url)
			pause := time.Duration(20+rnd.Intn(20)) * time.Second
			time.Sleep(pause)
			continue
		}

		for imgIdx, imgURL := range imageURLs {
			if err := downloadImageFile(httpClient, imgURL, outputDir, baseIndex, imgIdx); err != nil {
				fmt.Printf("[DEBUG] Playwright image download error base=%d idx=%d img=%d: %v\n", baseIndex, i, imgIdx, err)
				continue
			}
			downloaded++
		}

		pause := time.Duration(20+rnd.Intn(20)) * time.Second
		time.Sleep(pause)
	}

	fmt.Printf("[DEBUG] Playwright downloaded base=%d count=%d\n", baseIndex, downloaded)
	return downloaded
}

func extractImageURLs(page playwright.Page) ([]string, error) {
	locator := page.Locator("[data-marker=\"image-frame/image\"]")
	count, err := locator.Count()
	if err != nil {
		return nil, fmt.Errorf("locator count error: %w", err)
	}

	seen := make(map[string]bool)
	var urls []string
	for i := 0; i < count; i++ {
		src, err := locator.Nth(i).GetAttribute("src")
		if err != nil {
			continue
		}
		src = strings.TrimSpace(src)
		if src == "" {
			continue
		}
		if !seen[src] {
			seen[src] = true
			urls = append(urls, src)
		}
	}

	if len(urls) > 0 {
		return urls, nil
	}

	htmlContent, err := page.Content()
	if err != nil {
		return nil, fmt.Errorf("page content error: %w", err)
	}

	return extractImageURLsFromHTML(htmlContent)
}

func extractImageURLsFromHTML(content string) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("html parse error: %w", err)
	}

	seen := make(map[string]bool)
	var urls []string

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key == "data-marker" && attr.Val == "image-frame/image" {
					for _, a := range n.Attr {
						if a.Key == "src" {
							src := strings.TrimSpace(a.Val)
							if src != "" && !seen[src] {
								seen[src] = true
								urls = append(urls, src)
							}
							break
						}
					}
					break
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	return urls, nil
}

func downloadImageFile(client *http.Client, url, outputDir string, baseIndex, imgIdx int) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	}
	ua := userAgents[imgIdx%len(userAgents)]
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", "https://www.avito.ru/")
	req.Header.Set("Connection", "keep-alive")

	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			wait := jitteredBackoff(attempt, 5*time.Second)
			fmt.Printf("[DEBUG] Playwright image retry base=%d img=%d attempt=%d wait=%v url=%s\n", baseIndex, imgIdx, attempt, wait, url)
			time.Sleep(wait)
		}
		resp, err = client.Do(req)
		if err != nil {
			if attempt == 2 {
				return err
			}
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			_ = resp.Body.Close()
			continue
		}
		break
	}
	if resp == nil {
		return fmt.Errorf("nil response")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(body) < 100 {
		return fmt.Errorf("response too small: %d bytes", len(body))
	}

	ext := ".jpg"
	ct := resp.Header.Get("Content-Type")
	if ct != "" {
		if strings.Contains(ct, "webp") {
			ext = ".webp"
		} else if strings.Contains(ct, "png") {
			ext = ".png"
		} else if strings.Contains(ct, "jpeg") || strings.Contains(ct, "jpg") {
			ext = ".jpg"
		}
	}

	savePath := filepath.Join(outputDir, fmt.Sprintf("a%d_%d%s", baseIndex, imgIdx, ext))
	if err := os.WriteFile(savePath, body, 0644); err != nil {
		return err
	}

	pause := time.Duration(3+baseIndex%5) * time.Second
	time.Sleep(pause)
	return nil
}

func jitteredBackoff(attempt int, base time.Duration) time.Duration {
	if attempt <= 0 {
		return base
	}
	backoff := base * time.Duration(math.Pow(2, float64(attempt-1)))
	jitter := time.Duration(rand.Int63n(int64(base)))
	return backoff + jitter
}
