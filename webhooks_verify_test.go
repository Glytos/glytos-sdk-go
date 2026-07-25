package glytos

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
	"time"
)

// A fixed vector produced by the server signer (HMAC-SHA256 over "{ts}.{body}"),
// shared byte-for-byte with the Python and Node SDK test suites, so this SDK is
// proven compatible with how Glytos signs deliveries.
const (
	vecSecret = "whsec_test_secret"
	vecTS     = "1710000000"
	vecBody   = `{"event":"call.completed","id":"evt_123"}`
	vecSig    = "356d865446082d49c6bfa57299c45900ab0d6681363426341acbe7c8f07ba025"
)

func sigHeader(ts, sig string) string {
	return fmt.Sprintf("t=%s,v1=%s", ts, sig)
}

func TestVerifyWebhookAcceptsKnownSignature(t *testing.T) {
	// Tolerance 0 so the fixed historical timestamp is not rejected as too old.
	if !VerifyWebhook([]byte(vecBody), sigHeader(vecTS, vecSig), vecSecret, 0) {
		t.Fatal("expected the shared cross-SDK vector to verify")
	}
}

func TestVerifyWebhookRejectsTamperedBody(t *testing.T) {
	if VerifyWebhook([]byte(vecBody+"x"), sigHeader(vecTS, vecSig), vecSecret, 0) {
		t.Fatal("expected a tampered body to fail verification")
	}
}

func TestVerifyWebhookRejectsWrongSecret(t *testing.T) {
	if VerifyWebhook([]byte(vecBody), sigHeader(vecTS, vecSig), "wrong-secret", 0) {
		t.Fatal("expected a wrong secret to fail verification")
	}
}

func TestVerifyWebhookRejectsMalformedHeader(t *testing.T) {
	if VerifyWebhook([]byte(vecBody), "garbage", vecSecret, 0) {
		t.Fatal("expected a garbage header to fail verification")
	}
	if VerifyWebhook([]byte(vecBody), "t="+vecTS, vecSecret, 0) {
		t.Fatal("expected a header with no v1= to fail verification")
	}
}

func TestVerifyWebhookRejectsNonASCIISignature(t *testing.T) {
	// A non-ASCII v1= value must return false, not panic on the comparison.
	if VerifyWebhook([]byte(vecBody), sigHeader(vecTS, "deadbeefé"), vecSecret, 0) {
		t.Fatal("expected a non-ASCII signature to fail verification")
	}
}

func TestVerifyWebhookRejectsExpiredDelivery(t *testing.T) {
	// With the default tolerance the 2024 timestamp is far outside the window.
	if VerifyWebhook([]byte(vecBody), sigHeader(vecTS, vecSig), vecSecret, DefaultWebhookTolerance) {
		t.Fatal("expected an old delivery to be rejected under the default tolerance")
	}
}

func TestVerifyWebhookAcceptsFreshDelivery(t *testing.T) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	body := `{"hello":"world"}`
	secret := "whsec_live"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + body))
	sig := hex.EncodeToString(mac.Sum(nil))
	if !VerifyWebhook([]byte(body), sigHeader(ts, sig), secret, DefaultWebhookTolerance) {
		t.Fatal("expected a fresh, correctly signed delivery to verify")
	}
}

func TestWebhooksServiceVerifyDelegates(t *testing.T) {
	c := New("gly_test")
	if !c.Webhooks.Verify([]byte(vecBody), sigHeader(vecTS, vecSig), vecSecret, 0) {
		t.Fatal("Webhooks.Verify should delegate to VerifyWebhook and accept the vector")
	}
}
