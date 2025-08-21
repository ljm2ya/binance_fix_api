## Binance FIX API

A Go wrapper for Binance FIX API supporting both Order Entry and Market Data endpoints with ultra-low latency trading capabilities.

## Features

- **Order Entry**: Place and manage orders via `fix-oe.binance.com`
- **Market Data**: Subscribe to real-time trade streams via `fix-md.binance.com`
- **Built-in Configuration**: No external config files needed
- **Event-Driven**: Simple emission pattern for real-time data
- **Authentication**: ED25519 signature support
- **Ultra-Low Latency**: Optimized for high-frequency trading

## Requirements

- apiKey: Binance FIX API KEY
- privateKeyFilePath: Binance Ed25519 pem file path

## Quick Start

### Order Entry Example

```go
package main

import (
	"context"
	"time"

	fix "github.com/ljm2ya/binance_fix_api"
	"github.com/quickfixgo/enum"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()

	// Order Entry Configuration
	config := fix.Config{
		APIKey:             "your_api_key",
		PrivateKeyFilePath: "path/to/private_key.pem",
		Endpoint:           fix.OrderEntryEndpoint,
	}

	client, err := fix.NewClient(config, fix.WithZapLogFactory(logger.Sugar()))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		panic(err)
	}
	defer client.Stop()

	// Subscribe to execution reports
	client.SubscribeToExecutionReport(func(order *fix.Order) {
		logger.Info("Order update", zap.Any("order", order))
	})

	// Get account limits
	limit, err := client.NewGetLimitService().Do(ctx)
	if err != nil {
		panic(err)
	}
	logger.Info("Account limits", zap.Any("limits", limit))

	// Place a limit order
	order, err := client.NewOrderSingleService().
		Symbol("BTCUSDT").
		Side(enum.Side_BUY).
		Type(enum.OrdType_LIMIT).
		TimeInForce(enum.TimeInForce_GOOD_TILL_CANCEL).
		Quantity(0.001).
		Price(50000.00).
		Do(ctx)

	logger.Info("Order placed", zap.Any("order", order), zap.Error(err))

	time.Sleep(5 * time.Second)
}
```

### Market Data Example

```go
package main

import (
	"context"
	"time"

	fix "github.com/ljm2ya/binance_fix_api"
	"github.com/ljm2ya/binance_fix_api/handlers"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()

	// Market Data Configuration
	config := fix.Config{
		APIKey:             "your_api_key",
		PrivateKeyFilePath: "path/to/private_key.pem",
		Endpoint:           fix.MarketDataEndpoint,
	}

	client, err := fix.NewClient(config, fix.WithZapLogFactory(logger.Sugar()))
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		panic(err)
	}
	defer client.Stop()

	// Subscribe to trade stream
	client.SubscribeToTradeStream(func(trade *handlers.Trade) {
		logger.Info("Trade received",
			zap.String("symbol", trade.Symbol),
			zap.Int64("tradeID", trade.TradeID),
			zap.Float64("price", trade.Price),
			zap.Float64("quantity", trade.Quantity),
			zap.Time("time", trade.TradeTime),
		)
	})

	// Subscribe to multiple symbols
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	if err := client.SubscribeToTrades(ctx, symbols); err != nil {
		panic(err)
	}

	logger.Info("Subscribed to trade streams", zap.Strings("symbols", symbols))

	// Keep running to receive trades
	time.Sleep(30 * time.Second)
}
```

## Configuration

The library uses built-in endpoint configurations:

- **Order Entry**: `fix-oe.binance.com:9000` (SenderCompID: "BOETRADE")
- **Market Data**: `fix-md.binance.com:9000` (SenderCompID: "BMDWATCH")

No external config files required - everything is configured automatically based on the endpoint type.

## API Reference

### Client Methods

#### Order Entry
- `NewOrderSingleService()` - Create new single order
- `NewGetLimitService()` - Query account limits
- `SubscribeToExecutionReport(callback)` - Subscribe to order updates

#### Market Data
- `SubscribeToTrades(ctx, symbols)` - Subscribe to trade streams for multiple symbols
- `UnsubscribeFromTrades(ctx, symbols)` - Unsubscribe from trade streams
- `SubscribeToTradeStream(callback)` - Set trade stream callback handler

### Data Structures

#### Trade
```go
type Trade struct {
    Symbol        string    // Trading pair (e.g., "BTCUSDT")
    TradeID       int64     // Unique trade identifier
    Price         float64   // Trade price
    Quantity      float64   // Trade quantity
    TradeTime     time.Time // Trade timestamp
    BuyerOrderID  int64     // Buyer order ID
    SellerOrderID int64     // Seller order ID
    IsBuyerMaker  bool      // Whether buyer is maker
}
```

## Supported Messages

### Order Entry Messages
1. ✅ `NewOrderSingle<D>` - Submit new order
2. ✅ `ExecutionReport<8>` - Order state changes
3. ✅ `LimitQuery<XLQ>` - Query account limits
4. 🚫 `NewOrderList<E>` - Not implemented
5. 🚫 `OrderCancelRequest<F>` - Not implemented
6. 🚫 `OrderMassCancelRequest<q>` - Not implemented

### Market Data Messages
1. ✅ `MarketDataRequest<V>` - Subscribe to market data
2. ✅ `MarketDataIncrementalRefresh<X>` - Real-time trade data
3. ✅ `MarketDataSnapshotFullRefresh<W>` - Market data snapshots
