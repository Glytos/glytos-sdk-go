package glytos

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// DefaultWebhookTolerance is the default replay window, in seconds, for
// VerifyWebhook.
const DefaultWebhookTolerance = 300

// VerifyWebhook verifies a webhook delivery signature. Pass the RAW request
// body, the X-Glytos-Signature header value, and your endpoint secret. It
// returns true only if the signature is valid and, when toleranceSeconds > 0,
// the timestamp is within that many seconds of now.
//
// The scheme matches the server signer: HMAC-SHA256 over "{timestamp}.{body}",
// sent as "X-Glytos-Signature: t=<ts>,v1=<hex>". The comparison is
// constant-time and never panics on a malformed or non-ASCII signature.
func VerifyWebhook(payload []byte, signatureHeader, secret string, toleranceSeconds int) bool {
	var timestamp, provided string
	for _, piece := range strings.Split(signatureHeader, ",") {
		idx := strings.Index(piece, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(piece[:idx])
		value := strings.TrimSpace(piece[idx+1:])
		switch key {
		case "t":
			timestamp = value
		case "v1":
			provided = value
		}
	}
	if timestamp == "" || provided == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	// hmac.Equal is constant-time and tolerates unequal lengths, so a non-ASCII
	// or short v1= value returns false rather than panicking.
	if !hmac.Equal([]byte(expected), []byte(provided)) {
		return false
	}

	if toleranceSeconds > 0 {
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return false
		}
		delta := time.Now().Unix() - ts
		if delta < 0 {
			delta = -delta
		}
		if delta > int64(toleranceSeconds) {
			return false
		}
	}
	return true
}
