package fix

import (
	"bytes"
	"context"
	"errors"
	"strconv"

	"github.com/quickfixgo/quickfix"
)

var (
	ErrClosed             = errors.New("connection is closed")
	ErrInvalidRequestIDTag = errors.New("request id tag not found")
)

// call represents a FIX message call
type call struct {
	request  *quickfix.Message
	response *quickfix.Message
	done     chan error
}

// waiter wraps a call for waiting on response
type waiter struct {
	*call
}

// wait for the response message of an ongoing FIX call
func (w waiter) wait(ctx context.Context) (*quickfix.Message, error) {
	select {
	case err, ok := <-w.call.done:
		if !ok {
			err = ErrClosed
		}
		if err != nil {
			return nil, err
		}
		return w.call.response, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// copyMessage creates a deep copy of a FIX message
func copyMessage(msg *quickfix.Message) (*quickfix.Message, error) {
	out := quickfix.NewMessage()
	err := quickfix.ParseMessage(out, bytes.NewBufferString(msg.String()))
	if err != nil {
		return nil, err
	}
	return out, nil
}

// floatToString converts float64 to string with optimal precision
func floatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}