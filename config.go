package fix

import (
	"fmt"
	"strings"

	"github.com/quickfixgo/quickfix"
)

// EndpointType represents the type of FIX endpoint
type EndpointType string

const (
	OrderEntryEndpoint EndpointType = "OE"
	MarketDataEndpoint EndpointType = "MD"
)

// EndpointConfig contains endpoint-specific configuration
type EndpointConfig struct {
	Host           string
	Port           int
	SenderCompID   string
	TargetCompID   string
	HeartbeatInt   int
	ReconnectCount int
}

// DefaultEndpoints provides default Binance FIX endpoint configurations
var DefaultEndpoints = map[EndpointType]EndpointConfig{
	OrderEntryEndpoint: {
		Host:           "fix-oe.binance.com",
		Port:           9000,
		SenderCompID:   "BOETRADE", // BOE + TRADE
		TargetCompID:   "SPOT",
		HeartbeatInt:   30,
		ReconnectCount: 10,
	},
	MarketDataEndpoint: {
		Host:           "fix-md.binance.com",
		Port:           9000,
		SenderCompID:   "BMDWATCH", // BMD + WATCH
		TargetCompID:   "SPOT",
		HeartbeatInt:   30,
		ReconnectCount: 10,
	},
}

// GenerateQuickFixSettings creates QuickFIX settings from endpoint config
func GenerateQuickFixSettings(endpoint EndpointType, apiKey string, enableSSL bool) (*quickfix.Settings, error) {
	config, exists := DefaultEndpoints[endpoint]
	if !exists {
		return nil, fmt.Errorf("unknown endpoint type: %s", endpoint)
	}

	// Build settings string
	var settingsBuilder strings.Builder

	// Default section
	settingsBuilder.WriteString("[DEFAULT]\n")
	settingsBuilder.WriteString("BeginString=FIX.4.4\n")
	settingsBuilder.WriteString(fmt.Sprintf("SocketConnectHost=%s\n", config.Host))
	settingsBuilder.WriteString(fmt.Sprintf("SocketConnectPort=%d\n", config.Port))
	settingsBuilder.WriteString(fmt.Sprintf("HeartBtInt=%d\n", config.HeartbeatInt))
	settingsBuilder.WriteString(fmt.Sprintf("SenderCompID=%s\n", config.SenderCompID))
	settingsBuilder.WriteString(fmt.Sprintf("TargetCompID=%s\n", config.TargetCompID))
	settingsBuilder.WriteString("ConnectionType=initiator\n")
	//settingsBuilder.WriteString("ReconnectInterval=5\n")
	//settingsBuilder.WriteString("LogonTimeout=10\n")
	//settingsBuilder.WriteString("StartTime=00:00:00\n")
	//settingsBuilder.WriteString("EndTime=00:00:00\n")
	//settingsBuilder.WriteString("UseDataDictionary=N\n")
	//settingsBuilder.WriteString("ResetOnLogon=Y\n")
	//settingsBuilder.WriteString("ResetOnLogout=Y\n")
	//settingsBuilder.WriteString("ResetOnDisconnect=Y\n")
	//if config.ReconnectCount > 0 {
	//settingsBuilder.WriteString(fmt.Sprintf("MaxReconnectAttempts=%d\n", config.ReconnectCount))
	//}
	if enableSSL {
		settingsBuilder.WriteString("SocketUseSSL=Y\n")
		//settingsBuilder.WriteString("ValidateCertificates=Y\n")
	}
	settingsBuilder.WriteString("\n")

	// Session section
	settingsBuilder.WriteString("[SESSION]\n")
	//settingsBuilder.WriteString("BeginString=FIX.4.4\n")
	//settingsBuilder.WriteString(fmt.Sprintf("SenderCompID=%s\n", apiKey))
	//settingsBuilder.WriteString(fmt.Sprintf("TargetCompID=%s\n", config.TargetCompID))

	// Create settings from string
	settings, err := quickfix.ParseSettings(strings.NewReader(settingsBuilder.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	return settings, nil
}

// ConnectionConfig holds configuration for a FIX connection
type ConnectionConfig struct {
	Endpoint      EndpointType
	APIKey        string
	PrivateKeyPEM []byte
	EnableSSL     bool
}

// NewConnectionConfig creates a new connection configuration
func NewConnectionConfig(endpoint EndpointType, apiKey string, privateKeyPEM []byte) *ConnectionConfig {
	return &ConnectionConfig{
		Endpoint:      endpoint,
		APIKey:        apiKey,
		PrivateKeyPEM: privateKeyPEM,
		EnableSSL:     true, // Default to SSL enabled
	}
}
