package fix

import (
	"context"
	"crypto/ed25519"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chuckpreslar/emission"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/quickfix"
	"go.uber.org/zap"

	"github.com/ljm2ya/binance_fix_api/handlers"
)

const logonTimeout = 30 * time.Second

type Config struct {
	APIKey             string
	PrivateKeyFilePath string
	PrivateKeyPEM      []byte
	Settings           *quickfix.Settings
	Endpoint           EndpointType
}

type Options struct {
	messageHandling MessageHandling
	responseMode    ResponseMode
	fixLogFactory   quickfix.LogFactory
}


func defaultOpts() Options {
	return Options{
		messageHandling: MessageHandlingSequential,
		responseMode:    ResponseModeEverything,
		fixLogFactory:   quickfix.NewNullLogFactory(),
	}
}

type NewClientOption func(o *Options)

func WithMessageHandlingOpt(mh MessageHandling) NewClientOption {
	return func(o *Options) {
		o.messageHandling = mh
	}
}

func WithResponseModeOpt(rm ResponseMode) NewClientOption {
	return func(o *Options) {
		o.responseMode = rm
	}
}

func WithZapLogFactory(logger *zap.SugaredLogger) NewClientOption {
	return func(o *Options) {
		o.fixLogFactory = NewZapLogFactory(logger)
	}
}

func WithFixLogFactoryOpt(factory quickfix.LogFactory) NewClientOption {
	return func(o *Options) {
		o.fixLogFactory = factory
	}
}

type Client struct {
	mu          sync.Mutex
	isConnected atomic.Bool
	initiator   *quickfix.Initiator
	pending     map[string]*call
	emitter     *emission.Emitter

	apiKey       string
	privateKey   ed25519.PrivateKey
	beginString  string
	targetCompID string
	senderCompID string

	options Options
	config  Config  // Store original config for reconnection
}

func NewClient(conf Config, opts ...NewClientOption) (*Client, error) {
	// Generate settings if not provided
	var generatedSenderCompID string
	if conf.Settings == nil {
		var err error
		conf.Settings, generatedSenderCompID, err = GenerateQuickFixSettings(conf.Endpoint, conf.APIKey, true)
		if err != nil {
			return nil, err
		}
	}

	globalSettings := conf.Settings.GlobalSettings()
	beginString, err := globalSettings.Setting("BeginString")
	if err != nil {
		return nil, err
	}
	targetCompID, err := globalSettings.Setting("TargetCompID")
	if err != nil {
		return nil, err
	}
	senderCompID, err := globalSettings.Setting("SenderCompID")
	if err != nil {
		return nil, err
	}
	
	// Use generated SenderCompID if we created the settings
	if generatedSenderCompID != "" {
		senderCompID = generatedSenderCompID
	}

	var privateKey ed25519.PrivateKey
	if conf.PrivateKeyPEM != nil {
		privateKey, err = ParseEd25519PrivateKey(conf.PrivateKeyPEM)
		if err != nil {
			return nil, err
		}
	} else if conf.PrivateKeyFilePath != "" {
		privateKey, err = GetEd25519PrivateKeyFromFile(conf.PrivateKeyFilePath)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("either PrivateKeyPEM or PrivateKeyFilePath must be provided")
	}

	options := defaultOpts()
	for _, opt := range opts {
		opt(&options)
	}

	// Create a new Client object.
	client := &Client{
		pending:      make(map[string]*call),
		emitter:      emission.NewEmitter(),
		apiKey:       conf.APIKey,
		privateKey:   privateKey,
		beginString:  beginString,
		targetCompID: targetCompID,
		senderCompID: senderCompID,
		options:      options,
		config:       conf, // Store for reconnection
	}

	// Init session and logon to Binance FIX API server.
	client.initiator, err = quickfix.NewInitiator(
		client,
		quickfix.NewMemoryStoreFactory(),
		conf.Settings,
		options.fixLogFactory,
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Start(ctx context.Context) error {
	if err := c.initiator.Start(); err != nil {
		return err
	}

	// Wait for the session to be authorized by the server.
	timeoutCtx, cancel := context.WithTimeout(ctx, logonTimeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return errors.New("logon timed out")
		default:
			if c.IsConnected() {
				return nil
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (c *Client) IsConnected() bool {
	return c.isConnected.Load()
}

// SubscribeToDisconnect allows listening for disconnection events
func (c *Client) SubscribeToDisconnect(callback func(sessionID quickfix.SessionID)) {
	c.emitter.On("disconnect", func(args ...interface{}) {
		if len(args) > 0 {
			if sessionID, ok := args[0].(quickfix.SessionID); ok {
				callback(sessionID)
			}
		}
	})
}

// WaitForDisconnect blocks until the connection is lost (useful for long-running tests)
func (c *Client) WaitForDisconnect() <-chan bool {
	disconnected := make(chan bool, 1)
	c.SubscribeToDisconnect(func(_ quickfix.SessionID) {
		select {
		case disconnected <- true:
		default:
		}
	})
	return disconnected
}

// SubscribeToMaintenance allows listening for server maintenance notifications
func (c *Client) SubscribeToMaintenance(callback func(headline, text string)) {
	c.emitter.On("maintenance", func(args ...interface{}) {
		if len(args) > 0 {
			if newsData, ok := args[0].(map[string]string); ok {
				callback(newsData["headline"], newsData["text"])
			}
		}
	})
}

// SubscribeToReconnectNeeded allows listening for reconnection requirements
func (c *Client) SubscribeToReconnectNeeded(callback func()) {
	c.emitter.On("reconnect_needed", func(args ...interface{}) {
		callback()
	})
}

// WaitForMaintenanceOrDisconnect blocks until maintenance is announced or connection is lost
func (c *Client) WaitForMaintenanceOrDisconnect() <-chan string {
	events := make(chan string, 1)
	
	c.SubscribeToDisconnect(func(_ quickfix.SessionID) {
		select {
		case events <- "disconnect":
		default:
		}
	})
	
	c.SubscribeToMaintenance(func(headline, text string) {
		select {
		case events <- "maintenance":
		default:
		}
	})
	
	return events
}

// Stop closes underlying connection.
func (c *Client) Stop() {
	c.initiator.Stop()
}


// Call initiates a FIX call and wait for the response.
func (c *Client) Call(
	ctx context.Context, id string, msg *quickfix.Message,
) (*quickfix.Message, error) {
	call, err := c.send(id, msg)
	if err != nil {
		return nil, err
	}

	return call.wait(ctx)
}

// SendWithoutResponse sends a message without waiting for a response (for subscriptions)
func (c *Client) SendWithoutResponse(msg *quickfix.Message) error {
	if !c.isConnected.Load() {
		return ErrClosed
	}

	c.addCommonHeaders(msg)
	return quickfix.Send(msg)
}

func (c *Client) addCommonHeaders(msg *quickfix.Message) {
	msg.Header.Set(field.NewBeginString(c.beginString))
	msg.Header.Set(field.NewTargetCompID(c.targetCompID))
	msg.Header.Set(field.NewSenderCompID(c.senderCompID))
	msg.Header.Set(field.NewSendingTime(time.Now().UTC()))
}

func (c *Client) send(
	id string, msg *quickfix.Message,
) (waiter, error) {
	if !c.isConnected.Load() {
		return waiter{}, ErrClosed
	}

	c.addCommonHeaders(msg)
	cc := &call{request: msg, done: make(chan error, 1)}
	c.pending[id] = cc

	if err := quickfix.Send(msg); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return waiter{}, err
	}

	return waiter{cc}, nil
}

func (c *Client) handleSubscriptions(msgType string, msg *quickfix.Message) {
	if enum.MsgType(msgType) == enum.MsgType_EXECUTION_REPORT {
		order, err := handlers.DecodeExecutionReport(msg)
		if err != nil {
			return
		}
		c.emitter.Emit(ExecutionReportTopic, &order)
	} else if enum.MsgType(msgType) == enum.MsgType_MARKET_DATA_SNAPSHOT_FULL_REFRESH ||
		enum.MsgType(msgType) == enum.MsgType_MARKET_DATA_INCREMENTAL_REFRESH {
		trade, err := handlers.DecodeTradeMessage(msg)
		if err != nil {
			return
		}
		c.emitter.Emit(TradeStreamTopic, &trade)
	}
}

// handleNewsMessage processes News <B> messages for server maintenance notifications
func (c *Client) handleNewsMessage(msg *quickfix.Message) {
	// Extract news headline (Tag 148)
	headline := ""
	if msg.Body.Has(148) {
		headline, _ = msg.Body.GetString(148)
	}
	
	// Extract news text (Tag 58) 
	newsText := ""
	if msg.Body.Has(58) {
		newsText, _ = msg.Body.GetString(58)
	}
	
	// Check if this is a maintenance notification
	isMaintenanceNews := strings.Contains(strings.ToLower(headline), "maintenance") || 
		strings.Contains(strings.ToLower(newsText), "maintenance") ||
		strings.Contains(strings.ToLower(newsText), "reconnect")
	
	if isMaintenanceNews {
		// Emit maintenance event for applications to handle
		c.emitter.Emit("maintenance", map[string]string{
			"headline": headline,
			"text":     newsText,
		})
		
		// For Market Data connections, trigger reconnection logic
		if strings.Contains(c.senderCompID, "BMD") {
			c.emitter.Emit("reconnect_needed", true)
		}
	}
}
