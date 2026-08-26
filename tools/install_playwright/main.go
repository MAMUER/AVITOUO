package main

import (
	"fmt"
	"os"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	pw, err := playwright.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "playwright run error: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = pw.Stop() }()

	_, err = pw.Chromium.Launch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "chromium launch error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("OK")
}
