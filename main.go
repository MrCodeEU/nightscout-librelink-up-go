package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MrCodeEU/homelab-automation/apps/nightscout-librelink-up-go/config"
	"github.com/MrCodeEU/homelab-automation/apps/nightscout-librelink-up-go/librelink"
	"github.com/MrCodeEU/homelab-automation/apps/nightscout-librelink-up-go/nightscout"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

	log.Println("Starting Nightscout LibreLink Up Go Connector...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Configuration loaded - Region: %s, Interval: %d minutes, Nightscout URL: %s", 
		cfg.LinkUpRegion, cfg.LinkUpTimeInterval, cfg.NightscoutURL)

	// Create LibreLink client
	llClient, err := librelink.NewClient(cfg.LinkUpRegion, cfg.LinkUpUsername, cfg.LinkUpPassword)
	if err != nil {
		log.Fatalf("Failed to create LibreLink client: %v", err)
	}

	// Create Nightscout client
	nsClient := nightscout.NewClient(cfg.NightscoutURL, cfg.NightscoutAPIToken)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create ticker for polling interval
	ticker := time.NewTicker(time.Duration(cfg.LinkUpTimeInterval) * time.Minute)
	defer ticker.Stop()

	// Initial run
	if err := syncGlucoseData(llClient, nsClient); err != nil {
		log.Printf("Error during initial sync: %v", err)
	}

	// Main loop
	for {
		select {
		case <-ticker.C:
			if err := syncGlucoseData(llClient, nsClient); err != nil {
				log.Printf("Error syncing glucose data: %v", err)
			}
		case <-sigChan:
			log.Println("Shutdown signal received, exiting gracefully...")
			return
		}
	}
}

func syncGlucoseData(llClient *librelink.Client, nsClient *nightscout.Client) error {
	log.Println("Fetching glucose data from LibreLink Up...")

	// Authenticate with LibreLink
	if err := llClient.Login(); err != nil {
		return err
	}

	// Get connections (CGM sensors)
	connections, err := llClient.GetConnections()
	if err != nil {
		return err
	}

	if len(connections) == 0 {
		log.Println("No active LibreLink connections found")
		return nil
	}

	// Get latest glucose reading from first connection
	reading, err := llClient.GetLatestReading(connections[0].PatientID)
	if err != nil {
		return err
	}

	if reading == nil {
		log.Println("No glucose reading available")
		return nil
	}

	log.Printf("Glucose reading: %.1f %s (Trend: %s) at %s",
		reading.Value, reading.Unit, reading.TrendArrow, reading.Timestamp.Format(time.RFC3339))

	// Check if reading is recent (within last 15 minutes)
	if time.Since(reading.Timestamp) > 15*time.Minute {
		log.Printf("Warning: Reading is %v old, may be stale", time.Since(reading.Timestamp))
	}

	// Post to Nightscout
	if err := nsClient.PostGlucoseReading(reading); err != nil {
		return err
	}

	log.Println("Successfully posted glucose data to Nightscout")
	return nil
}
