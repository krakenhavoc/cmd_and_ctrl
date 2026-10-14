package github

// app.go: filing as a GitHub App (#3493), so an in-app report shows the
// app's bot as its author instead of the maintainer whose personal
// access token used to file it.
//
// A GitHub App authenticates in two steps. It signs a short-lived JWT
// with its private key (RS256, issuer = the app id), and trades that for
// an installation access token for the one repository it is installed
// on. The installation token is what the issue calls send, and it lasts
// an hour, so it is cached and swapped for a fresh one shortly before it
// runs out.
//
// Standard library only, like the rest of the package: RS256 is
// crypto/rsa over a SHA-256 digest, and a JWT is three base64url parts.
// Neither the private key, the JWT nor the installation token ever
// appears in a log line, an error string or a response.

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
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// TokenSource supplies the bearer token for each API call.
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

// staticToken is a personal access token: the same value every call.
type staticToken string

func (s staticToken) Token(context.Context) (string, error) { return string(s), nil }

const (
	// jwtLifetime is how long the app JWT is valid. GitHub's ceiling is
	// ten minutes; the iat is backdated a minute for clock drift.
	jwtLifetime = 9 * time.Minute
	jwtBackdate = time.Minute
	// refreshBefore swaps an installation token for a fresh one this
	// long before GitHub says it expires.
	refreshBefore = 5 * time.Minute
)

// AppTokenSource mints and caches installation tokens for one
// installation of one GitHub App.
type AppTokenSource struct {
	appID          int64
	installationID int64
	key            *rsa.PrivateKey
	baseURL        string
	http           *http.Client
	now            func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

// NewAppTokenSource returns a token source for the app and installation,
// signing with the PEM private key GitHub issued for the app. The key
// may also arrive base64-encoded on one line (an env file cannot hold a
// multi-line PEM).
func NewAppTokenSource(appID, installationID int64, privateKey []byte) (*AppTokenSource, error) {
	if appID <= 0 || installationID <= 0 {
		return nil, errors.New("github: app id and installation id must be positive")
	}
	key, err := ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return &AppTokenSource{
		appID:          appID,
		installationID: installationID,
		key:            key,
		baseURL:        defaultBaseURL,
		http:           &http.Client{Timeout: defaultTimeout},
		now:            time.Now,
	}, nil
}

// ParsePrivateKey reads an RSA private key: PEM (PKCS#1, which GitHub
// issues, or PKCS#8), or that PEM base64-encoded. The error never quotes
// the input.
func ParsePrivateKey(raw []byte) (*rsa.PrivateKey, error) {
	text := strings.TrimSpace(string(raw))
	if !strings.HasPrefix(text, "-----BEGIN") {
		decoded, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return nil, errors.New("github: app private key is neither PEM nor base64-encoded PEM")
		}
		text = string(decoded)
	}
	block, _ := pem.Decode([]byte(text))
	if block == nil {
		return nil, errors.New("github: app private key has no PEM block")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("github: app private key is not an RSA key in PKCS#1 or PKCS#8")
	}
	k, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("github: app private key is not an RSA key")
	}
	return k, nil
}

// jwt signs the app's JWT for one token exchange.
func (a *AppTokenSource) jwt() (string, error) {
	now := a.now()
	enc := base64.RawURLEncoding
	header := enc.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claims, err := json.Marshal(struct {
		IAT int64  `json:"iat"`
		EXP int64  `json:"exp"`
		ISS string `json:"iss"`
	}{
		IAT: now.Add(-jwtBackdate).Unix(),
		EXP: now.Add(jwtLifetime).Unix(),
		ISS: fmt.Sprint(a.appID),
	})
	if err != nil {
		return "", err
	}
	signing := header + "." + enc.EncodeToString(claims)
	digest := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, a.key, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("github: sign app jwt: %w", err)
	}
	return signing + "." + enc.EncodeToString(sig), nil
}

// Token returns a live installation token, minting one when the cached
// one is missing or about to expire.
func (a *AppTokenSource) Token(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && a.now().Before(a.expires.Add(-refreshBefore)) {
		return a.token, nil
	}
	jwt, err := a.jwt()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/app/installations/%d/access_tokens", a.baseURL, a.installationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", fmt.Errorf("github: build token request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", userAgent)
	resp, err := a.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("github: installation token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	lim := io.LimitReader(resp.Body, maxResponseBytes)
	if resp.StatusCode != http.StatusCreated {
		slurp, _ := io.ReadAll(lim)
		return "", &APIError{Op: "installation token", Status: resp.StatusCode, Body: truncate(string(slurp), 300)}
	}
	var out struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(lim).Decode(&out); err != nil {
		return "", fmt.Errorf("github: decode installation token: %w", err)
	}
	if out.Token == "" {
		return "", errors.New("github: installation token response carried no token")
	}
	a.token, a.expires = out.Token, out.ExpiresAt
	return a.token, nil
}

// NewAppClient returns a Client for "owner/name" that authenticates as
// the app installation behind src.
func NewAppClient(src *AppTokenSource, repo string) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		repo:    strings.Trim(repo, "/"),
		tokens:  src,
		http:    &http.Client{Timeout: defaultTimeout},
	}
}
