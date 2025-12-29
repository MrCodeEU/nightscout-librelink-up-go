package nightscout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MrCodeEU/homelab-automation/apps/nightscout-librelink-up-go/librelink"
)

// Client represents a Nightscout API client
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// Entry represents a Nightscout glucose entry
type Entry struct {
	Type       string  `json:"type"`
	SGV        float64 `json:"sgv"`
	Direction  string  `json:"direction"`
	Device     string  `json:"device"`
	Date       int64   `json:"date"`
	DateString string  `json:"dateString"`
}

// NewClient creates a new Nightscout API client
func NewClient(baseURL, apiToken string) *Client {
	return &Client{
		baseURL:  baseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PostGlucoseReading posts a glucose reading to Nightscout
func (c *Client) PostGlucoseReading(reading *librelink.GlucoseReading) error {
	if reading == nil {
		return fmt.Errorf("reading is nil")
	}

	entry := Entry{
		Type:       "sgv",
		SGV:        reading.Value,
		Direction:  convertTrendArrow(reading.TrendArrow),
		Device:     "nightscout-librelink-up-go",
		Date:       reading.Timestamp.UnixMilli(),
		DateString: reading.Timestamp.Format(time.RFC3339),
	}

	entries := []Entry{entry}
	reqBody, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	// Add http:// or https:// if not present
	baseURL := c.baseURL
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		// Use http:// for internal Docker hostnames with ports (e.g., nightscout:1337)
		// Use https:// for public domains (e.g., ns.mljr.eu)
		if strings.Contains(baseURL, ":") {
			baseURL = "http://" + baseURL
		} else {
			baseURL = "https://" + baseURL
		}
	}

	url := fmt.Sprintf("%s/api/v1/entries", baseURL)
	log.Printf("Posting glucose data to: %s", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-secret", c.apiToken)
	// Trick Nightscout into thinking this is a secure connection to avoid redirects
	req.Header.Set("X-Forwarded-Proto", "https")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// convertTrendArrow converts LibreLink trend arrows to Nightscout direction format
func convertTrendArrow(arrow string) string {
	switch arrow {
	case "DoubleUp":
		return "DoubleUp"
	case "SingleUp":
		return "SingleUp"
	case "FortyFiveUp":
		return "FortyFiveUp"
	case "Flat":
		return "Flat"
	case "FortyFiveDown":
		return "FortyFiveDown"
	case "SingleDown":
		return "SingleDown"
	case "DoubleDown":
		return "DoubleDown"
	default:
		return "NONE"
	}
}
