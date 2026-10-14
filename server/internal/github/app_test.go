package github

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// testAppKey is one 2048-bit key for the whole file: generating one per
// test would cost more than the tests.
var testAppKey = func() *rsa.PrivateKey {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return k
}()

func pkcs1PEM() []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(testAppKey)})
}

// verifyJWT checks an RS256 JWT against testAppKey and returns its claims.
func verifyJWT(t *testing.T, jwt string) map[string]any {
	t.Helper()
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		t.Fatalf("jwt has %d parts", len(parts))
	}
	enc := base64.RawURLEncoding
	sig, err := enc.DecodeString(parts[2])
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(&testAppKey.PublicKey, crypto.SHA256, digest[:], sig); err != nil {
		t.Fatalf("jwt signature: %v", err)
	}
	head, _ := enc.DecodeString(parts[0])
	if !strings.Contains(string(head), `"alg":"RS256"`) {
		t.Errorf("jwt header = %s", head)
	}
	raw, _ := enc.DecodeString(parts[1])
	var claims map[string]any
	if err := json.Unmarshal(raw, &claims); err != nil {
		t.Fatal(err)
	}
	return claims
}

// appServer fakes the two calls an app client makes: the token exchange
// for installation 42, and filing an issue with the token it handed out.
type appServer struct {
	exchanges atomic.Int32
	expiresIn time.Duration
	refuse    bool
	srv       *httptest.Server
}

func newAppServer(t *testing.T) *appServer {
	a := &appServer{expiresIn: time.Hour}
	a.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/app/installations/42/access_tokens":
			claims := verifyJWT(t, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
			if claims["iss"] != "7" {
				t.Errorf("iss = %v, want the app id", claims["iss"])
			}
			if a.refuse {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"message":"A JSON web token could not be decoded"}`))
				return
			}
			n := a.exchanges.Add(1)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"token":      fmt.Sprintf("inst-%d", n),
				"expires_at": time.Now().Add(a.expiresIn).UTC().Format(time.RFC3339),
			})
		case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/issues":
			if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer inst-") {
				t.Errorf("issue filed with %q, want an installation token", r.Header.Get("Authorization"))
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"number":1,"html_url":"https://github.com/o/r/issues/1"}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(a.srv.Close)
	return a
}

func appClient(t *testing.T, a *appServer, key []byte) (*Client, *AppTokenSource) {
	t.Helper()
	src, err := NewAppTokenSource(7, 42, key)
	if err != nil {
		t.Fatal(err)
	}
	c := NewAppClient(src, "o/r").WithBaseURL(a.srv.URL).WithHTTPClient(a.srv.Client())
	return c, src
}

func TestAppClientFilesWithAnInstallationTokenAndCachesIt(t *testing.T) {
	a := newAppServer(t)
	c, _ := appClient(t, a, pkcs1PEM())
	for i := 0; i < 3; i++ {
		if _, _, err := c.CreateIssue(context.Background(), "t", "b", nil); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}
	if n := a.exchanges.Load(); n != 1 {
		t.Errorf("token exchanges = %d, want 1 (cached)", n)
	}
}

func TestAppTokenIsRefreshedBeforeItExpires(t *testing.T) {
	a := newAppServer(t)
	c, src := appClient(t, a, pkcs1PEM())
	now := time.Now()
	src.now = func() time.Time { return now }
	if _, _, err := c.CreateIssue(context.Background(), "t", "b", nil); err != nil {
		t.Fatal(err)
	}
	// Inside the refresh margin of the hour: a fresh token.
	now = now.Add(time.Hour - refreshBefore + time.Second)
	if _, _, err := c.CreateIssue(context.Background(), "t", "b", nil); err != nil {
		t.Fatal(err)
	}
	if n := a.exchanges.Load(); n != 2 {
		t.Errorf("token exchanges = %d, want 2", n)
	}
}

func TestAppJWTClaims(t *testing.T) {
	src, err := NewAppTokenSource(7, 42, pkcs1PEM())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_800_000_000, 0)
	src.now = func() time.Time { return now }
	jwt, err := src.jwt()
	if err != nil {
		t.Fatal(err)
	}
	claims := verifyJWT(t, jwt)
	if claims["iat"] != float64(now.Add(-jwtBackdate).Unix()) || claims["exp"] != float64(now.Add(jwtLifetime).Unix()) {
		t.Errorf("claims = %v", claims)
	}
	if exp := now.Add(jwtLifetime).Sub(now.Add(-jwtBackdate)); exp > 10*time.Minute {
		t.Errorf("jwt spans %s, past GitHub's ten minutes", exp)
	}
}

func TestParsePrivateKeyFormats(t *testing.T) {
	pkcs8, err := x509.MarshalPKCS8PrivateKey(testAppKey)
	if err != nil {
		t.Fatal(err)
	}
	for name, raw := range map[string][]byte{
		"pkcs1":        pkcs1PEM(),
		"pkcs8":        pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}),
		"base64 pkcs1": []byte(base64.StdEncoding.EncodeToString(pkcs1PEM())),
		"padded":       append(append([]byte("\n  "), pkcs1PEM()...), '\n'),
	} {
		k, err := ParsePrivateKey(raw)
		if err != nil || !k.Equal(testAppKey) {
			t.Errorf("%s: %v", name, err)
		}
	}
	secret := "not-a-key-SECRETVALUE"
	if _, err := ParsePrivateKey([]byte(secret)); err == nil || strings.Contains(err.Error(), "SECRETVALUE") {
		t.Errorf("garbage: %v", err)
	}
}

func TestNewAppTokenSourceRefusesBadIDs(t *testing.T) {
	if _, err := NewAppTokenSource(0, 42, pkcs1PEM()); err == nil {
		t.Error("app id 0 accepted")
	}
	if _, err := NewAppTokenSource(7, -1, pkcs1PEM()); err == nil {
		t.Error("installation id -1 accepted")
	}
}

func TestARefusedTokenExchangeIsAnAPIErrorWithNoCredential(t *testing.T) {
	a := newAppServer(t)
	a.refuse = true
	c, src := appClient(t, a, pkcs1PEM())
	jwt, _ := src.jwt()
	_, _, err := c.CreateIssue(context.Background(), "t", "b", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode() != http.StatusUnauthorized {
		t.Fatalf("err = %v, want a 401 APIError", err)
	}
	if strings.Contains(err.Error(), jwt[:20]) || strings.Contains(err.Error(), "PRIVATE KEY") {
		t.Errorf("the error carries a credential: %v", err)
	}
}
