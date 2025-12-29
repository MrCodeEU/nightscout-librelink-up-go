package librelink

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const (
	appVersion  = "4.17.0"
	appProduct  = "llu.ios"
	contentType = "application/json"
)

// Regional API endpoints for LibreLink Up
var endpoints = map[string]string{
	"AE":  "https://api-ae.libreview.io",
	"AP":  "https://api-ap.libreview.io",
	"AU":  "https://api-au.libreview.io",
	"CA":  "https://api-ca.libreview.io",
	"DE":  "https://api-de.libreview.io",
	"EU":  "https://api-eu.libreview.io",
	"EU2": "https://api-eu2.libreview.io",
	"FR":  "https://api-fr.libreview.io",
	"JP":  "https://api-jp.libreview.io",
	"US":  "https://api-us.libreview.io",
	"LA":  "https://api-la.libreview.io",
	"RU":  "https://api-ru.libreview.io",
	"CN":  "https://api-cn.libreview.io",
}

// Client represents a LibreLink Up API client
type Client struct {
	baseURL    string
	username   string
	password   string
	authToken  string
	accountID  string
	httpClient *http.Client
}

// Connection represents a LibreLink sensor connection
type Connection struct {
	PatientID   string `json:"patientId"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	SensorState int    `json:"sensor"`
}

// GlucoseReading represents a blood glucose measurement
type GlucoseReading struct {
	Value      float64
	Unit       string
	Timestamp  time.Time
	TrendArrow string
}

// API request/response structures
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Status int `json:"status"`
	Data   struct {
		User struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
		AuthTicket struct {
			Token     string `json:"token"`
			ExpiresAt int64  `json:"expires"`
		} `json:"authTicket"`
	} `json:"data"`
}

type connectionsResponse struct {
	Status int `json:"status"`
	Data   []struct {
		PatientID string `json:"patientId"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Sensor    struct {
			DeviceID string `json:"deviceId"`
			Serial   string `json:"sn"`
		} `json:"sensor"`
	} `json:"data"`
}

type glucoseResponse struct {
	Status int `json:"status"`
	Data   struct {
		Connection struct {
			GlucoseMeasurement struct {
				Value          float64 `json:"Value"`
				ValueInMgPerDl float64 `json:"ValueInMgPerDl"`
				TrendArrow     int     `json:"TrendArrow"`
				Timestamp      string  `json:"Timestamp"`
			} `json:"glucoseMeasurement"`
		} `json:"connection"`
	} `json:"data"`
}

// NewClient creates a new LibreLink Up client
func NewClient(region, username, password string) (*Client, error) {
	baseURL, ok := endpoints[region]
	if !ok {
		return nil, fmt.Errorf("unsupported region: %s", region)
	}

	// Create cookie jar to maintain session state (like withCredentials in TypeScript)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}, nil
}

// Login authenticates with LibreLink Up and retrieves auth token
func (c *Client) Login() error {
	loginReq := loginRequest{
		Email:    c.username,
		Password: c.password,
	}

	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	fmt.Printf("DEBUG: Attempting login to %s with email: %s\n", c.baseURL, c.username)
	fmt.Printf("DEBUG: Request body: %s\n", string(reqBody))

	req, err := http.NewRequest("POST", c.baseURL+"/llu/auth/login", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	c.setHeaders(req)

	fmt.Println("DEBUG: Request headers:")
	for k, v := range req.Header {
		fmt.Printf("  %s: %s\n", k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with HTTP status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp loginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	if loginResp.Status != 0 {
		fmt.Printf("DEBUG: Full login response: %+v\n", loginResp)
		return fmt.Errorf("login failed with API status: %d", loginResp.Status)
	}

	c.authToken = loginResp.Data.AuthTicket.Token
	c.accountID = loginResp.Data.User.ID
	fmt.Printf("DEBUG: Successfully authenticated, token: %s...\n", c.authToken[:20])
	fmt.Printf("DEBUG: Account ID: %s\n", c.accountID)
	return nil
}

// GetConnections retrieves all LibreLink connections (sensors)
func (c *Client) GetConnections() ([]Connection, error) {
	if c.authToken == "" {
		return nil, fmt.Errorf("not authenticated, call Login() first")
	}

	// Always try to fetch connections list
	req, err := http.NewRequest("GET", c.baseURL+"/llu/connections", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create connections request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Authorization", "Bearer "+c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connections request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("DEBUG: Connections response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Connections response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("connections request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var connResp connectionsResponse
	if err := json.Unmarshal(body, &connResp); err != nil {
		return nil, fmt.Errorf("failed to decode connections response: %w", err)
	}

	// If no connections found, patient account - return account ID as connection
	if len(connResp.Data) == 0 {
		fmt.Printf("DEBUG: No connections found - using patient account ID as connection: %s\n", c.accountID)
		return []Connection{
			{
				PatientID: c.accountID,
				FirstName: "Self",
				LastName:  "",
			},
		}, nil
	}

	connections := make([]Connection, len(connResp.Data))
	for i, conn := range connResp.Data {
		connections[i] = Connection{
			PatientID: conn.PatientID,
			FirstName: conn.FirstName,
			LastName:  conn.LastName,
		}
	}

	return connections, nil
}

// GetLatestReading retrieves the latest glucose reading for a patient
func (c *Client) GetLatestReading(patientID string) (*GlucoseReading, error) {
	if c.authToken == "" {
		return nil, fmt.Errorf("not authenticated, call Login() first")
	}

	url := fmt.Sprintf("%s/llu/connections/%s/graph", c.baseURL, patientID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create glucose request: %w", err)
	}

	c.setHeaders(req)
	req.Header.Set("Authorization", "Bearer "+c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("glucose request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("DEBUG: Graph response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Graph response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("glucose request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var glucoseResp glucoseResponse
	if err := json.Unmarshal(body, &glucoseResp); err != nil {
		return nil, fmt.Errorf("failed to decode glucose response: %w", err)
	}

	measurement := glucoseResp.Data.Connection.GlucoseMeasurement

	// Parse timestamp (format: "11/19/2024 3:14:29 PM")
	timestamp, err := parseLibreLinkTimestamp(measurement.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return &GlucoseReading{
		Value:      measurement.ValueInMgPerDl,
		Unit:       "mg/dL",
		Timestamp:  timestamp,
		TrendArrow: trendArrowToString(measurement.TrendArrow),
	}, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", contentType)
	req.Header.Set("product", appProduct)
	req.Header.Set("version", appVersion)

	// Add SHA-256 hashed account-id if authenticated
	if c.accountID != "" && c.authToken != "" {
		hasher := sha256.New()
		hasher.Write([]byte(c.accountID))
		hashedAccountID := hex.EncodeToString(hasher.Sum(nil))
		req.Header.Set("account-id", hashedAccountID)
	}
}

func parseLibreLinkTimestamp(timestamp string) (time.Time, error) {
	// Try parsing with common formats
	formats := []string{
		"1/2/2006 3:04:05 PM",
		"01/02/2006 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		// Parse in local timezone (container timezone)
		if t, err := time.ParseInLocation(format, timestamp, time.Local); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestamp)
}

func trendArrowToString(arrow int) string {
	switch arrow {
	case 1:
		return "DoubleUp"
	case 2:
		return "SingleUp"
	case 3:
		return "FortyFiveUp"
	case 4:
		return "Flat"
	case 5:
		return "FortyFiveDown"
	case 6:
		return "SingleDown"
	case 7:
		return "DoubleDown"
	default:
		return "Unknown"
	}
}
