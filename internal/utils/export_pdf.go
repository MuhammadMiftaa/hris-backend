package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"hris-backend/config/env"
)

// GeneratePDF — render HTML string ke PDF menggunakan Gotenberg
func GeneratePDF(htmlContent string) ([]byte, error) {
	gotenbergURL := env.Cfg.Gotenberg.URL
	if gotenbergURL == "" {
		gotenbergURL = "http://gotenberg:3000"
	}
	endpoint := fmt.Sprintf("%s/forms/chromium/convert/html", gotenbergURL)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Gotenberg mengharapkan file bernama "index.html"
	part, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("gotenberg: create form file: %w", err)
	}
	if _, err := io.WriteString(part, htmlContent); err != nil {
		return nil, fmt.Errorf("gotenberg: write html: %w", err)
	}

	// Paper size A4
	if err := writer.WriteField("paperWidth", "8.27"); err != nil {
		return nil, fmt.Errorf("gotenberg: write field paperWidth: %w", err)
	}
	if err := writer.WriteField("paperHeight", "11.69"); err != nil {
		return nil, fmt.Errorf("gotenberg: write field paperHeight: %w", err)
	}

	// Margin (inch)
	if err := writer.WriteField("marginTop", "0.5"); err != nil {
		return nil, fmt.Errorf("gotenberg: write field marginTop: %w", err)
	}
	if err := writer.WriteField("marginBottom", "0.5"); err != nil {
		return nil, fmt.Errorf("gotenberg: write field marginBottom: %w", err)
	}
	if err := writer.WriteField("marginLeft", "0.5"); err != nil {
		return nil, fmt.Errorf("gotenberg: write field marginLeft: %w", err)
	}
	if err := writer.WriteField("marginRight", "0.5"); err != nil {
		return nil, fmt.Errorf("gotenberg: write field marginRight: %w", err)
	}

	// Print background graphics (CSS background-color, dsb.)
	if err := writer.WriteField("printBackground", "true"); err != nil {
		return nil, fmt.Errorf("gotenberg: write field printBackground: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("gotenberg: close writer: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(http.MethodPost, endpoint, &body)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gotenberg: status %d — %s", resp.StatusCode, string(errBody))
	}

	pdf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: read response: %w", err)
	}

	return pdf, nil
}