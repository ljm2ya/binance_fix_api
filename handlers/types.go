package handlers

import "github.com/quickfixgo/enum"

// Order status types
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusPendingCancel   OrderStatus = "PENDING_CANCEL"
	OrderStatusRejected        OrderStatus = "REJECTED"
	OrderStatusPendingNew      OrderStatus = "PENDING_NEW"
	OrderStatusExpired         OrderStatus = "EXPIRED"
)

var mappedOrderStatus = map[enum.OrdStatus]OrderStatus{
	enum.OrdStatus_NEW:              OrderStatusNew,
	enum.OrdStatus_PARTIALLY_FILLED: OrderStatusPartiallyFilled,
	enum.OrdStatus_FILLED:           OrderStatusFilled,
	enum.OrdStatus_CANCELED:         OrderStatusCanceled,
	enum.OrdStatus_PENDING_CANCEL:   OrderStatusPendingCancel,
	enum.OrdStatus_REJECTED:         OrderStatusRejected,
	enum.OrdStatus_PENDING_NEW:      OrderStatusPendingNew,
	enum.OrdStatus_EXPIRED:          OrderStatusExpired,
}

// Time in force types
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GOOD_TILL_CANCEL"
	TimeInForceIOC TimeInForce = "IMMEDIATE_OR_CANCEL"
	TimeInForceFOK TimeInForce = "FILL_OR_KILL"
)

var mappedTimeInForce = map[enum.TimeInForce]TimeInForce{
	enum.TimeInForce_GOOD_TILL_CANCEL:    TimeInForceGTC,
	enum.TimeInForce_IMMEDIATE_OR_CANCEL: TimeInForceIOC,
	enum.TimeInForce_FILL_OR_KILL:        TimeInForceFOK,
}

// Order type types
type OrderType string

const (
	OrderTypeMarket    OrderType = "MARKET"
	OrderTypeLimit     OrderType = "LIMIT"
	OrderTypeStop      OrderType = "STOP"
	OrderTypeStopLimit OrderType = "STOP_LIMIT"
)

var mappedOrderType = map[enum.OrdType]OrderType{
	enum.OrdType_MARKET:     OrderTypeMarket,
	enum.OrdType_LIMIT:      OrderTypeLimit,
	enum.OrdType_STOP:       OrderTypeStop,
	enum.OrdType_STOP_LIMIT: OrderTypeStopLimit,
}

// Side types
type SideType string

const (
	SideTypeBuy  SideType = "BUY"
	SideTypeSell SideType = "SELL"
)

var mappedSideType = map[enum.Side]SideType{
	enum.Side_BUY:  SideTypeBuy,
	enum.Side_SELL: SideTypeSell,
}