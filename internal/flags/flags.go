//revive:disable:package-comments
package flags

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// AgentConfig define agent config structure.
type AgentConfig struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

// ServerConfig define server config structure.
type ServerConfig struct {
	Address string
}

// GetEnvOrFlag function for getting flag or env for strings.
func GetEnvOrFlag(envVar string, flagValue string, defaultValue string) string {
	if value, ok := os.LookupEnv(envVar); ok {
		return value
	}
	if flagValue != "" {
		return flagValue
	}
	return defaultValue
}

// GetIntEnvOrFlag function for getting flag or env for numbers.
func GetIntEnvOrFlag(envVar string, flagValue int, defaultValue int) int {
	if value, ok := os.LookupEnv(envVar); ok {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("%s: %s", "Invalid flag value", envVar)
	}
	if flagValue != 0 {
		return flagValue
	}
	return defaultValue
}

// logErrorf error function.
func logErrorf(v ...interface{}) error {
	const stdErr = 2
	format := ""
	if len(v) > 0 {
		if str, ok := v[0].(string); ok {
			format = str
			v = v[1:]
		}
	}
	err := log.Output(stdErr, "ERROR: "+fmt.Sprintf(format, v...))
	if err != nil {
		return fmt.Errorf("%s: %w", "Failed write to log", err)
	}
	return nil
}

// ParseAgentConfig function of parsing agent config.
func ParseAgentConfig() (AgentConfig, error) {
	var (
		flagAddr      string
		flagReportInt int
		flagPollInt   int
	)

	const (
		defaultflagAddr  string = "localhost:8080"
		defaultReportInt int    = 10
		defaultPollInt   int    = 2
	)

	flag.StringVar(&flagAddr, "a", defaultflagAddr, "HTTP server endpoint address")
	flag.IntVar(&flagReportInt, "r", defaultReportInt, "Report interval in seconds")
	flag.IntVar(&flagPollInt, "p", defaultPollInt, "Poll interval in seconds")

	flag.Parse()

	if len(flag.Args()) > 0 {
		return AgentConfig{}, logErrorf("Unknown flags provided: %v", flag.Args())
	}

	address := GetEnvOrFlag("ADDRESS", flagAddr, defaultflagAddr)
	reportInterval := time.Duration(GetIntEnvOrFlag("REPORT_INTERVAL", flagReportInt, defaultReportInt)) * time.Second
	pollInterval := time.Duration(GetIntEnvOrFlag("POLL_INTERVAL", flagPollInt, defaultPollInt)) * time.Second

	return AgentConfig{
		Address:        address,
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
	}, nil
}

// ParseServerConfig function of parsing server config.
func ParseServerConfig() (ServerConfig, error) {
	var flagAddr string
	const defaultflagAddr string = "localhost:8080"

	flag.StringVar(&flagAddr, "a", defaultflagAddr, "HTTP server address")

	flag.Parse()

	if len(flag.Args()) > 0 {
		return ServerConfig{}, logErrorf("Unknown flags provided: %v", flag.Args())
	}

	address := GetEnvOrFlag("ADDRESS", flagAddr, defaultflagAddr)

	return ServerConfig{
		Address: address,
	}, nil
}
