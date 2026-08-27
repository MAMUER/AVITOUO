package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	_ "github.com/mattn/go-sqlite3"
)

const chromeUserDataDir = "User Data"

type dataBlob struct {
	cbData uint32
	pbData *byte
}

func decryptDPAPI(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	crypt32 := syscall.NewLazyDLL("crypt32.dll")
	proc := crypt32.NewProc("CryptUnprotectData")

	var outBlob dataBlob
	r, _, err := proc.Call(
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(0),
		uintptr(0),
		uintptr(0),
		uintptr(0),
		0,
		uintptr(unsafe.Pointer(&outBlob)),
	)

	if r == 0 {
		return nil, fmt.Errorf("CryptUnprotectData failed: %v", err)
	}

	defer func() { _, _ = syscall.LocalFree(syscall.Handle(uintptr(unsafe.Pointer(outBlob.pbData)))) }()

	result := make([]byte, outBlob.cbData)
	for i := uint32(0); i < outBlob.cbData; i++ {
		result[i] = *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(outBlob.pbData)) + uintptr(i)))
	}

	return result, nil
}

func decryptChromeCookie(encrypted []byte, key []byte) ([]byte, error) {
	if len(encrypted) < 3 {
		return nil, fmt.Errorf("encrypted data too short")
	}

	prefix := string(encrypted[:3])
	if prefix == "v10" || prefix == "v11" {
		if len(encrypted) < 15 {
			return nil, fmt.Errorf("encrypted data too short for v10/v11")
		}

		nonce := encrypted[3:15]
		ciphertext := encrypted[15 : len(encrypted)-16]
		tag := encrypted[len(encrypted)-16:]

		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}

		aesGCM, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}

		plaintext, err := aesGCM.Open(nil, nonce, append(ciphertext, tag...), nil)
		if err != nil {
			return nil, err
		}

		return plaintext, nil
	}

	return decryptDPAPI(encrypted)
}

func getChromeMasterKey() ([]byte, error) {
	localStatePath := filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", chromeUserDataDir, "Local State")

	data, err := os.ReadFile(localStatePath)
	if err != nil {
		return nil, err
	}

	var localState struct {
		OsCrypt struct {
			EncryptedKey string `json:"encrypted_key"`
		} `json:"os_crypt"`
	}

	if err := json.Unmarshal(data, &localState); err != nil {
		return nil, err
	}

	if localState.OsCrypt.EncryptedKey == "" {
		return nil, fmt.Errorf("no encrypted_key found")
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(localState.OsCrypt.EncryptedKey)
	if err != nil {
		return nil, err
	}

	if len(encryptedKey) < 5 || string(encryptedKey[:5]) != "DPAPI" {
		return nil, fmt.Errorf("invalid encrypted_key prefix")
	}

	encryptedKey = encryptedKey[5:]

	return decryptDPAPI(encryptedKey)
}

func readChromeCookies(domain, cookiesPath string) (map[string]string, error) {
	masterKey, err := getChromeMasterKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get master key: %v", err)
	}

	roURI := "file:" + cookiesPath + "?mode=ro"
	db, err := sql.Open("sqlite3", roURI)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query(`
		SELECT name, value, encrypted_value 
		FROM cookies 
		WHERE host_key LIKE ? OR host_key LIKE ? OR host_key = ?
	`, "%."+domain, domain, "."+domain)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	cookies := make(map[string]string)
	for rows.Next() {
		var name, value string
		var encryptedValue []byte
		if rows.Scan(&name, &value, &encryptedValue) != nil {
			continue
		}

		if len(encryptedValue) > 0 {
			decrypted, err := decryptChromeCookie(encryptedValue, masterKey)
			if err == nil {
				cookies[name] = string(decrypted)
				continue
			}
		}

		if value != "" {
			cookies[name] = value
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chrome cookies scan error: %w", err)
	}

	return cookies, nil
}

func readFirefoxCookies(domain, cookiesPath string) (map[string]string, error) {
	roURI := "file:" + cookiesPath + "?mode=ro"
	db, err := sql.Open("sqlite3", roURI)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query("SELECT name, value FROM moz_cookies WHERE host LIKE ? OR host LIKE ? OR host = ?",
		"%."+domain, domain, "."+domain)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	cookies := make(map[string]string)
	for rows.Next() {
		var name, value string
		if rows.Scan(&name, &value) != nil {
			continue
		}
		cookies[name] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("firefox cookies scan error: %w", err)
	}

	return cookies, nil
}

func findFirefoxCookies(domain string) (map[string]string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return nil, fmt.Errorf("APPDATA not set")
	}

	profilesDir := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		cookiesPath := filepath.Join(profilesDir, entry.Name(), "cookies.sqlite")
		if _, err := os.Stat(cookiesPath); err != nil {
			continue
		}

		cookies, err := readFirefoxCookies(domain, cookiesPath)
		if err == nil && len(cookies) > 0 {
			return cookies, nil
		}
	}

	return nil, fmt.Errorf("no Firefox cookies found")
}

func findChromeCookies(domain string) (map[string]string, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return nil, fmt.Errorf("LOCALAPPDATA not set")
	}

	chromeBase := filepath.Join(localAppData, "Google", "Chrome", chromeUserDataDir)

	profiles := []string{"Default", "Profile 1", "Profile 2", "Profile 3", "Profile 4", "Profile 5", "Profile 6", "Profile 7", "Profile 8", "Profile 9"}

	for _, profile := range profiles {
		cookiesPath := filepath.Join(chromeBase, profile, "Network", "Cookies")
		if _, err := os.Stat(cookiesPath); err != nil {
			continue
		}

		cookies, err := readChromeCookies(domain, cookiesPath)
		if err == nil && len(cookies) > 0 {
			return cookies, nil
		}
	}

	return nil, fmt.Errorf("no Chrome cookies found")
}

func findEdgeCookies(domain string) (map[string]string, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return nil, fmt.Errorf("LOCALAPPDATA not set")
	}

	edgeBase := filepath.Join(localAppData, "Microsoft", "Edge", chromeUserDataDir)

	profiles := []string{"Default", "Profile 1", "Profile 2", "Profile 3", "Profile 4", "Profile 5", "Profile 6", "Profile 7", "Profile 8", "Profile 9"}

	for _, profile := range profiles {
		cookiesPath := filepath.Join(edgeBase, profile, "Network", "Cookies")
		if _, err := os.Stat(cookiesPath); err != nil {
			continue
		}

		cookies, err := readChromeCookies(domain, cookiesPath)
		if err == nil && len(cookies) > 0 {
			return cookies, nil
		}
	}

	return nil, fmt.Errorf("no Edge cookies found")
}

// GetBrowserCookies tries to get cookies from Firefox, Chrome, or Edge for the given domain
func GetBrowserCookies(domain string) (map[string]string, error) {
	if cookies, err := findFirefoxCookies(domain); err == nil && len(cookies) > 0 {
		fmt.Printf("[DEBUG] Loaded %d cookies from Firefox for %s\n", len(cookies), domain)
		return cookies, nil
	}

	if cookies, err := findChromeCookies(domain); err == nil && len(cookies) > 0 {
		fmt.Printf("[DEBUG] Loaded %d cookies from Chrome for %s\n", len(cookies), domain)
		return cookies, nil
	}

	if cookies, err := findEdgeCookies(domain); err == nil && len(cookies) > 0 {
		fmt.Printf("[DEBUG] Loaded %d cookies from Edge for %s\n", len(cookies), domain)
		return cookies, nil
	}

	return nil, fmt.Errorf("no browser cookies found for %s", domain)
}
