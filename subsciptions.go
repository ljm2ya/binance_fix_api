package fix

import "github.com/ljm2ya/binance_fix_api/handlers"

type ExecutionReportHandler func(o *handlers.Order)

func (c *Client) SubscribeToExecutionReport(listener ExecutionReportHandler) {
	c.emitter.On(ExecutionReportTopic, listener)
}

type TradeStreamHandler func(trade *handlers.Trade)

func (c *Client) SubscribeToTradeStream(listener TradeStreamHandler) {
	c.emitter.On(TradeStreamTopic, listener)
}
