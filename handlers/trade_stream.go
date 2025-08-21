package handlers

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/quickfixgo/field"
	"github.com/quickfixgo/quickfix"
)

// Trade represents a trade from the market data stream
type Trade struct {
	Symbol      string
	TradeID     int64
	Price       float64
	Quantity    float64
	TradeTime   time.Time
	BuyerOrderID  int64
	SellerOrderID int64
	IsBuyerMaker  bool
}

// TradeStreamHandler manages trade data subscriptions
type TradeStreamHandler struct {
	subscriptions map[string][]TradeCallback
	mu            sync.RWMutex
}

// TradeCallback defines the function signature for trade callbacks
type TradeCallback func(trade Trade)

// NewTradeStreamHandler creates a new trade stream handler
func NewTradeStreamHandler() *TradeStreamHandler {
	return &TradeStreamHandler{
		subscriptions: make(map[string][]TradeCallback),
	}
}

// Subscribe adds a callback for trade updates on a specific symbol
func (h *TradeStreamHandler) Subscribe(symbol string, callback TradeCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.subscriptions[symbol] = append(h.subscriptions[symbol], callback)
}

// Unsubscribe removes all callbacks for a symbol
func (h *TradeStreamHandler) Unsubscribe(symbol string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	delete(h.subscriptions, symbol)
}

// HandleTradeMessage processes incoming trade messages and notifies subscribers
func (h *TradeStreamHandler) HandleTradeMessage(msg *quickfix.Message) error {
	trade, err := DecodeTradeMessage(msg)
	if err != nil {
		return err
	}

	h.mu.RLock()
	callbacks, exists := h.subscriptions[trade.Symbol]
	h.mu.RUnlock()

	if !exists {
		return nil // No subscribers for this symbol
	}

	// Notify all subscribers in parallel for ultra-low latency
	for _, callback := range callbacks {
		go callback(trade)
	}

	return nil
}

// DecodeTradeMessage parses a FIX trade message into a Trade struct
func DecodeTradeMessage(msg *quickfix.Message) (Trade, error) {
	symbol, err := getTradeSymbol(msg)
	if err != nil {
		return Trade{}, err
	}

	tradeID, err := getTradeID(msg)
	if err != nil {
		return Trade{}, err
	}

	price, err := getTradePrice(msg)
	if err != nil {
		return Trade{}, err
	}

	quantity, err := getTradeQuantity(msg)
	if err != nil {
		return Trade{}, err
	}

	tradeTime, err := getTradeTime(msg)
	if err != nil {
		return Trade{}, err
	}

	buyerOrderID, _ := getBuyerOrderID(msg)
	sellerOrderID, _ := getSellerOrderID(msg)
	isBuyerMaker, _ := getIsBuyerMaker(msg)

	return Trade{
		Symbol:        symbol,
		TradeID:       tradeID,
		Price:         price,
		Quantity:      quantity,
		TradeTime:     tradeTime,
		BuyerOrderID:  buyerOrderID,
		SellerOrderID: sellerOrderID,
		IsBuyerMaker:  isBuyerMaker,
	}, nil
}

// Trade field extraction functions optimized for performance

func getTradeSymbol(msg *quickfix.Message) (string, error) {
	var f field.SymbolField
	if err := msg.Body.Get(&f); err != nil {
		return "", err
	}
	return f.Value(), nil
}

func getTradeID(msg *quickfix.Message) (int64, error) {
	// Using TradeID field (Tag 1003) for Binance
	if msg.Body.Has(1003) {
		str, err := msg.Body.GetString(1003)
		if err != nil {
			return 0, err
		}
		return strconv.ParseInt(str, 10, 64)
	}
	// Fallback to TradeReportID field (Tag 571)
	if msg.Body.Has(571) {
		str, err := msg.Body.GetString(571)
		if err != nil {
			return 0, err
		}
		return strconv.ParseInt(str, 10, 64)
	}
	return 0, errors.New("trade ID not found")
}

func getTradePrice(msg *quickfix.Message) (float64, error) {
	// Use MDEntryPx field (Tag 270) for market data
	if msg.Body.Has(270) {
		str, err := msg.Body.GetString(270)
		if err != nil {
			return 0, err
		}
		return strconv.ParseFloat(str, 64)
	}
	// Fallback to LastPx field (Tag 31)
	var f field.LastPxField
	if msg.Body.Has(f.Tag()) {
		if err := msg.Body.Get(&f); err != nil {
			return 0, err
		}
		return f.InexactFloat64(), nil
	}
	return 0, errors.New("trade price not found")
}

func getTradeQuantity(msg *quickfix.Message) (float64, error) {
	// Use MDEntrySize field (Tag 271) for market data
	if msg.Body.Has(271) {
		str, err := msg.Body.GetString(271)
		if err != nil {
			return 0, err
		}
		return strconv.ParseFloat(str, 64)
	}
	// Fallback to LastQty field (Tag 32)
	var f field.LastQtyField
	if msg.Body.Has(f.Tag()) {
		if err := msg.Body.Get(&f); err != nil {
			return 0, err
		}
		return f.InexactFloat64(), nil
	}
	return 0, errors.New("trade quantity not found")
}

func getTradeTime(msg *quickfix.Message) (time.Time, error) {
	// Use TransactTime field (Tag 60)
	var f field.TransactTimeField
	if msg.Body.Has(f.Tag()) {
		if err := msg.Body.Get(&f); err != nil {
			return time.Time{}, err
		}
		return f.Value(), nil
	}
	
	return time.Time{}, errors.New("trade time not found")
}

func getBuyerOrderID(msg *quickfix.Message) (int64, error) {
	// Custom tag for buyer order ID (may vary by exchange)
	if msg.Body.Has(6010) {
		str, err := msg.Body.GetString(6010)
		if err != nil {
			return 0, err
		}
		return strconv.ParseInt(str, 10, 64)
	}
	return 0, nil
}

func getSellerOrderID(msg *quickfix.Message) (int64, error) {
	// Custom tag for seller order ID (may vary by exchange)
	if msg.Body.Has(6011) {
		str, err := msg.Body.GetString(6011)
		if err != nil {
			return 0, err
		}
		return strconv.ParseInt(str, 10, 64)
	}
	return 0, nil
}

func getIsBuyerMaker(msg *quickfix.Message) (bool, error) {
	// Custom tag for buyer maker flag (may vary by exchange)
	if msg.Body.Has(6012) {
		str, err := msg.Body.GetString(6012)
		if err != nil {
			return false, err
		}
		return str == "true" || str == "Y" || str == "1", nil
	}
	return false, nil
}