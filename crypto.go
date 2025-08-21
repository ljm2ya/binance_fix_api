package fix

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/quickfixgo/enum"
)

const (
	utcTimestampMillisFmt = "20060102-15:04:05.000"
	blockTypePrivateKey   = "PRIVATE KEY"
)

var (
	ErrNilPrivateKeyValue = errors.New("nil private key value")
	ErrInvalidEd25519Key  = errors.New("invalid key ed25519 key")
)

// ParseEd25519PrivateKey parses ED25519 private key from PEM data
func ParseEd25519PrivateKey(data []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != blockTypePrivateKey {
		return nil, ErrNilPrivateKeyValue
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	ret, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, ErrInvalidEd25519Key
	}

	return ret, nil
}

// GetEd25519PrivateKeyFromFile loads ED25519 private key from file
func GetEd25519PrivateKeyFromFile(path string) (ed25519.PrivateKey, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return ParseEd25519PrivateKey(data)
}

// GetLogonRawData creates authentication signature for FIX logon
func GetLogonRawData(
	privateKey ed25519.PrivateKey,
	senderCompID, targetCompID, sendingTime string,
) string {
	method := string(enum.MsgType_LOGON)
	msgSeqNum := "1" // Logon is the first request of fix protocol.
	payload := strings.Join([]string{method, senderCompID, targetCompID, msgSeqNum, sendingTime}, "\x01")
	data := ed25519.Sign(privateKey, []byte(payload))

	return base64.StdEncoding.EncodeToString(data)
}

// SendingTimeNow returns current UTC timestamp in FIX format
func SendingTimeNow() string {
	return time.Now().UTC().Format(utcTimestampMillisFmt)
}