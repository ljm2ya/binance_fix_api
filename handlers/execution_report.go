package handlers

import (
	"errors"
	"strconv"
	"time"

	"github.com/quickfixgo/field"
	"github.com/quickfixgo/quickfix"
)

const (
	utcTimestampMicrosFmt = "20060102-15:04:05.000000"
	tagCumQuoteQty        = 381
	tagOrderCreationTime  = 6635
	tagWorkingTime        = 636
)

// Order represents a trading order with all relevant fields
type Order struct {
	Symbol            string
	OrderID           int64
	ClientOrderID     string
	Price             float64
	OrderQty          float64
	CumQty            float64
	CumQuoteQty       float64
	Status            OrderStatus
	TimeInForce       TimeInForce
	Type              OrderType
	Side              SideType
	IcebergQuantity   float64
	TransactTime      time.Time
	OrderCreationTime time.Time
	WorkingTime       time.Time
}

// DecodeExecutionReport parses a FIX ExecutionReport message into an Order struct
func DecodeExecutionReport(msg *quickfix.Message) (Order, error) {
	status, err := getOrderStatus(msg)
	if err != nil {
		return Order{}, err
	}

	if status == OrderStatusRejected {
		reason, err := getText(msg)
		if err != nil {
			return Order{}, err
		}
		if reason != "" {
			return Order{}, errors.New(reason)
		}
	}

	symbol, err := getSymbol(msg)
	if err != nil {
		return Order{}, err
	}

	orderID, err := getOrderID(msg)
	if err != nil {
		return Order{}, err
	}

	clientOrderID, err := getClientOrderID(msg)
	if err != nil {
		return Order{}, err
	}

	price, err := getPrice(msg)
	if err != nil {
		return Order{}, err
	}

	orderQty, err := getOrderQty(msg)
	if err != nil {
		return Order{}, err
	}

	cumQty, err := getCumQty(msg)
	if err != nil {
		return Order{}, err
	}

	cumQuoteQty, err := getCumQuoteQty(msg)
	if err != nil {
		return Order{}, err
	}

	timeInForce, err := getTimeInForce(msg)
	if err != nil {
		return Order{}, err
	}

	orderType, err := getOrdType(msg)
	if err != nil {
		return Order{}, err
	}

	side, err := getSide(msg)
	if err != nil {
		return Order{}, err
	}

	maxFloor, err := getMaxFloor(msg)
	if err != nil {
		return Order{}, err
	}

	transactTime, err := getTransactTime(msg)
	if err != nil {
		return Order{}, err
	}

	orderCreationTime, err := getOrderCreationTime(msg)
	if err != nil {
		return Order{}, err
	}

	workingTime, err := getWorkingTime(msg)
	if err != nil {
		return Order{}, err
	}

	return Order{
		Symbol:            symbol,
		OrderID:           orderID,
		ClientOrderID:     clientOrderID,
		Price:             price,
		OrderQty:          orderQty,
		CumQty:            cumQty,
		CumQuoteQty:       cumQuoteQty,
		Status:            status,
		TimeInForce:       timeInForce,
		Type:              orderType,
		Side:              side,
		IcebergQuantity:   maxFloor,
		TransactTime:      transactTime,
		OrderCreationTime: orderCreationTime,
		WorkingTime:       workingTime,
	}, nil
}

// Field extraction functions

func getText(msg *quickfix.Message) (v string, err error) {
	var f field.TextField
	if msg.Body.Has(f.Tag()) {
		if err = msg.Body.Get(&f); err == nil {
			v = f.Value()
		}
	}
	return
}

func getSymbol(msg *quickfix.Message) (v string, err error) {
	var f field.SymbolField
	if err = msg.Body.Get(&f); err == nil {
		v = f.Value()
	}
	return
}

func getOrderID(msg *quickfix.Message) (v int64, err error) {
	var f field.OrderIDField
	if msg.Body.Has(f.Tag()) {
		if err = msg.Body.Get(&f); err != nil {
			return
		}
	}

	return strconv.ParseInt(f.Value(), 10, 64)
}

func getClientOrderID(msg *quickfix.Message) (v string, err error) {
	var f field.ClOrdIDField
	if msg.Body.Has(f.Tag()) {
		if err = msg.Body.Get(&f); err == nil {
			v = f.Value()
		}
	}
	return
}

func getOrderStatus(msg *quickfix.Message) (v OrderStatus, err error) {
	var f field.OrdStatusField
	if err = msg.Body.Get(&f); err == nil {
		v = mappedOrderStatus[f.Value()]
	}
	return
}

func getOrdType(msg *quickfix.Message) (v OrderType, err error) {
	var f field.OrdTypeField
	if err = msg.Body.Get(&f); err == nil {
		v = mappedOrderType[f.Value()]
	}
	return
}

func getSide(msg *quickfix.Message) (v SideType, err error) {
	var f field.SideField
	if err = msg.Body.Get(&f); err == nil {
		v = mappedSideType[f.Value()]
	}
	return
}

func getTimeInForce(msg *quickfix.Message) (v TimeInForce, err error) {
	var f field.TimeInForceField
	if msg.Body.Has(f.Tag()) {
		if err = msg.Body.Get(&f); err == nil {
			v = mappedTimeInForce[f.Value()]
		}
	}
	return
}

func getPrice(msg *quickfix.Message) (float64, error) {
	var f field.PriceField
	if msg.Body.Has(f.Tag()) {
		if err := msg.Body.Get(&f); err != nil {
			return 0, err
		}
		return f.InexactFloat64(), nil
	}
	return 0, nil
}

func getOrderQty(msg *quickfix.Message) (float64, error) {
	var f field.OrderQtyField
	if msg.Body.Has(f.Tag()) {
		if err := msg.Body.Get(&f); err != nil {
			return 0, err
		}
		return f.InexactFloat64(), nil
	}
	return 0, nil
}

func getCumQty(msg *quickfix.Message) (float64, error) {
	var f field.CumQtyField
	if msg.Body.Has(f.Tag()) {
		if err := msg.Body.Get(&f); err != nil {
			return 0, err
		}
		return f.InexactFloat64(), nil
	}
	return 0, nil
}

func getCumQuoteQty(msg *quickfix.Message) (float64, error) {
	if msg.Body.Has(tagCumQuoteQty) {
		str, err := msg.Body.GetString(tagCumQuoteQty)
		if err != nil {
			return 0, err
		}
		return strconv.ParseFloat(str, 64)
	}
	return 0, nil
}

func getMaxFloor(msg *quickfix.Message) (float64, error) {
	var f field.MaxFloorField
	if msg.Body.Has(f.Tag()) {
		if err := msg.Body.Get(&f); err != nil {
			return 0, err
		}
		return f.InexactFloat64(), nil
	}
	return 0, nil
}

func getTransactTime(msg *quickfix.Message) (v time.Time, err error) {
	var f field.TransactTimeField
	if msg.Body.Has(f.Tag()) {
		if err = msg.Body.Get(&f); err == nil {
			v = f.Value()
		}
	}
	return
}

func getOrderCreationTime(msg *quickfix.Message) (time.Time, error) {
	if msg.Body.Has(tagOrderCreationTime) {
		str, err := msg.Body.GetString(tagOrderCreationTime)
		if err != nil {
			return time.Time{}, err
		}
		return time.Parse(utcTimestampMicrosFmt, str)
	}
	return time.Time{}, nil
}

func getWorkingTime(msg *quickfix.Message) (time.Time, error) {
	if msg.Body.Has(tagWorkingTime) {
		str, err := msg.Body.GetString(tagWorkingTime)
		if err != nil {
			return time.Time{}, err
		}
		return time.Parse(utcTimestampMicrosFmt, str)
	}
	return time.Time{}, nil
}