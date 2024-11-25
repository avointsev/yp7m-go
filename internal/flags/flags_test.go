package flags

import (
	"bytes"
	"flag"
	"testing"
	"time"
)

func TestAgentEnvVariables(t *testing.T) {
	t.Setenv("ADDRESS", "envhost:9090")
	t.Setenv("REPORT_INTERVAL", "15")
	t.Setenv("POLL_INTERVAL", "5")

	flagAddr := "flaghost:8081"
	flagReportInt := 10
	flagPollInt := 2

	address := GetEnvOrFlag("ADDRESS", flagAddr, "localhost:8080")
	reportInterval := time.Duration(GetIntEnvOrFlag("REPORT_INTERVAL", flagReportInt, 10)) * time.Second
	pollInterval := time.Duration(GetIntEnvOrFlag("POLL_INTERVAL", flagPollInt, 2)) * time.Second

	if address != "envhost:9090" {
		t.Errorf("Expected address to be 'envhost:9090' from environment variable, got %s", address)
	}
	if reportInterval != 15*time.Second {
		t.Errorf("Expected reportInterval to be 15 seconds from environment variable, got %v", reportInterval)
	}
	if pollInterval != 5*time.Second {
		t.Errorf("Expected pollInterval to be 5 seconds from environment variable, got %v", pollInterval)
	}
}

func TestAgentFlagDefaults(t *testing.T) {
	testFlags := flag.NewFlagSet("test_flags", flag.ExitOnError)
	var testFlagAddr string
	var testFlagReportInt, testFlagPollInt int

	// Define flags in the new FlagSet
	testFlags.StringVar(&testFlagAddr, "a", "localhost:8080", "HTTP server endpoint address")
	testFlags.IntVar(&testFlagReportInt, "r", 10, "Report interval in seconds")
	testFlags.IntVar(&testFlagPollInt, "p", 2, "Poll interval in seconds")

	// Parse the flags (no custom values are provided, so defaults are used)
	if err := testFlags.Parse([]string{}); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	if testFlagAddr != "localhost:8080" {
		t.Errorf("Expected default address to be 'localhost:8080', got %s", testFlagAddr)
	}
	if testFlagReportInt != 10 {
		t.Errorf("Expected default report interval to be 10, got %d", testFlagReportInt)
	}
	if testFlagPollInt != 2 {
		t.Errorf("Expected default poll interval to be 2, got %d", testFlagPollInt)
	}
}

func TestAgentInvalidFlags(t *testing.T) {
	var (
		buf           bytes.Buffer
		flagAddr      string
		flagReportInt int
		flagPollInt   int
	)

	// Create a new FlagSet for testing unknown flags
	testFlags := flag.NewFlagSet("test_invalid_flags", flag.ContinueOnError)
	testFlags.SetOutput(&buf)

	// Define flags in the new FlagSet
	testFlags.StringVar(&flagAddr, "a", "localhost:8080", "HTTP server endpoint address")
	testFlags.IntVar(&flagReportInt, "r", 10, "Report interval in seconds")
	testFlags.IntVar(&flagPollInt, "p", 2, "Poll interval in seconds")

	// Parse an unknown flag
	err := testFlags.Parse([]string{"-unknown"})

	// Check if error contains "unknown flag" warning
	if err == nil || !bytes.Contains(buf.Bytes(), []byte("flag provided but not defined")) {
		t.Error("Expected error message for unknown flag")
	}
}

func TestServerEnvVariables(t *testing.T) {
	t.Setenv("ADDRESS", "envhost:9090")
	flagAddr := "flaghost:8081"

	address := GetEnvOrFlag("ADDRESS", flagAddr, "localhost:8080")

	if address != "envhost:9090" {
		t.Errorf("Expected address to be 'envhost:9090' from environment variable, got %s", address)
	}
}

func TestServerFlagDefaults(t *testing.T) {
	testFlags := flag.NewFlagSet("test_flags", flag.ExitOnError)
	var testFlagAddr string

	testFlags.StringVar(&testFlagAddr, "a", "localhost:8080", "HTTP server endpoint address")

	if err := testFlags.Parse([]string{}); err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}

	if testFlagAddr != "localhost:8080" {
		t.Errorf("Expected default address to be 'localhost:8080', got %s", testFlagAddr)
	}
}

func TestServerInvalidFlags(t *testing.T) {
	var (
		buf      bytes.Buffer
		flagAddr string
	)

	testFlags := flag.NewFlagSet("test_invalid_flags", flag.ContinueOnError)
	testFlags.SetOutput(&buf)

	testFlags.StringVar(&flagAddr, "a", "localhost:8080", "HTTP server endpoint address")
	err := testFlags.Parse([]string{"-unknown"})

	if err == nil || !bytes.Contains(buf.Bytes(), []byte("flag provided but not defined")) {
		t.Error("Expected error message for unknown flag")
	}
}
