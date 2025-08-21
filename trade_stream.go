package fix

import (
	"context"
	"fmt"
	"time"

	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/quickfix"
)


// SubscribeToTrades subscribes to trade data for specified symbols
func (c *Client) SubscribeToTrades(ctx context.Context, symbols []string) error {
	// Create market data request
	msg := quickfix.NewMessage()
	msg.Header.Set(field.NewMsgType(enum.MsgType_MARKET_DATA_REQUEST))
	
	// Generate unique request ID
	mdReqID := fmt.Sprintf("MDR_%d", time.Now().UnixNano())
	msg.Body.Set(field.NewMDReqID(mdReqID))
	msg.Body.Set(field.NewSubscriptionRequestType(enum.SubscriptionRequestType_SNAPSHOT_PLUS_UPDATES))
	msg.Body.Set(field.NewMarketDepth(1)) // Only trade data
	
	// Add symbols to request
	noRelatedSymGroup := quickfix.NewRepeatingGroup(146, // NoRelatedSym
		quickfix.GroupTemplate{quickfix.GroupElement(55)}) // Symbol
	
	for _, symbol := range symbols {
		group := noRelatedSymGroup.Add()
		group.Set(field.NewSymbol(symbol))
	}
	
	msg.Body.SetGroup(noRelatedSymGroup)
	
	// Add entry types (only trade data)
	noMDEntryTypesGroup := quickfix.NewRepeatingGroup(267, // NoMDEntryTypes
		quickfix.GroupTemplate{quickfix.GroupElement(269)}) // MDEntryType
	
	tradeGroup := noMDEntryTypesGroup.Add()
	tradeGroup.Set(field.NewMDEntryType(enum.MDEntryType_TRADE))
	msg.Body.SetGroup(noMDEntryTypesGroup)

	// Send request (no response expected for subscriptions)
	return c.SendWithoutResponse(msg)
}

// UnsubscribeFromTrades unsubscribes from trade data for specified symbols
func (c *Client) UnsubscribeFromTrades(ctx context.Context, symbols []string) error {
	// Create unsubscribe request
	msg := quickfix.NewMessage()
	msg.Header.Set(field.NewMsgType(enum.MsgType_MARKET_DATA_REQUEST))
	
	mdReqID := fmt.Sprintf("MDR_UNSUB_%d", time.Now().UnixNano())
	msg.Body.Set(field.NewMDReqID(mdReqID))
	msg.Body.Set(field.NewSubscriptionRequestType(enum.SubscriptionRequestType_DISABLE_PREVIOUS_SNAPSHOT_PLUS_UPDATE_REQUEST))

	// Add symbols to unsubscribe
	noRelatedSymGroup := quickfix.NewRepeatingGroup(146, // NoRelatedSym
		quickfix.GroupTemplate{quickfix.GroupElement(55)}) // Symbol
	
	for _, symbol := range symbols {
		group := noRelatedSymGroup.Add()
		group.Set(field.NewSymbol(symbol))
	}
	
	msg.Body.SetGroup(noRelatedSymGroup)

	// Send unsubscribe request (no response expected)
	return c.SendWithoutResponse(msg)
}