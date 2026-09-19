package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Discord OAuth2 + API endpoints. Split out as package-level vars
// rather than consts so tests can point them at an httptest.Server
// without touching the call sites.
var (
	authorizeEndpoint = "https://discord.com/oauth2/authorize"
	tokenEndpoint     = "https://discord.com/api/oauth2/token"
	userEndpoint      = "https://discord.com/api/users/@me"
)

// discordScopes is the set we ask for on every authorize. `identify`
// is the minimum to read the user's username, avatar, and global
// display name via /users/@me. We deliberately do not ask for
// `email` or `guilds` — the hobby use case doesn't need them, and
// asking for the minimum reduces the user-facing consent scope
// (the Discord prompt reads "will see your username and avatar",
// not "will see your email address").
const discordScopes = "identify"

// RequestedScopes is the scope string every authorize asks for. The
// callback records it on the identity when Discord's token response
// leaves `scope` out.
const RequestedScopes = discordScopes

// AuthorizeURL builds the URL the client should be redirected to.
// state + codeChallenge come from StateStore.Start; the caller
// handles the HTTP 302 themselves so they can attach cookies /
// set response headers as needed.
//
// prompt=none would skip Discord's consent screen on repeat
// sign-ins, but we leave it out — requesting consent every time
// makes the "scope of access" transparent to users who might not
// remember granting it.
func (c Config) AuthorizeURL(state, codeChallenge string) string {
	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", c.ClientID)
	q.Set("redirect_uri", c.RedirectURI)
	q.Set("scope", discordScopes)
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	return authorizeEndpoint + "?" + q.Encode()
}

// TokenResponse is the subset of Discord's token-exchange response
// we care about. The access token is used once, to read /users/@me,
// and then dropped; the server mints its own session. The refresh
// token is kept (ADR 0051 decision 5, S34): the callback hands it to
// the user store, which stores it encrypted under
// CMDCTRL_IDENTITY_KEY, or discards it when that key is not set. It
// is never logged and never sent to the client.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

// TokenErrorResponse mirrors Discord's RFC 6749 error body
// (error + error_description). Surfaced verbatim in the callback
// handler's 4xx so a misconfigured redirect_uri is obvious in
// the browser rather than collapsed to "invalid code".
type TokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// ExchangeCode trades the authorization code Discord just bounced
// back for an access token. codeVerifier is the PKCE value we
// parked in StateStore at the start of the flow.
func (c Config) ExchangeCode(ctx context.Context, client *http.Client, code, codeVerifier string) (TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", c.RedirectURI)
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)
	form.Set("code_verifier", codeVerifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("discord: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("discord: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("discord: read token body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var errResp TokenErrorResponse
		_ = json.Unmarshal(body, &errResp)
		if errResp.Error != "" {
			return TokenResponse{}, fmt.Errorf("discord: token exchange: %s (%s)", errResp.Error, errResp.ErrorDescription)
		}
		return TokenResponse{}, fmt.Errorf("discord: token exchange: status %d", resp.StatusCode)
	}
	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return TokenResponse{}, fmt.Errorf("discord: parse token body: %w", err)
	}
	return tr, nil
}

// User is the subset of Discord's /users/@me response we store.
// GlobalName is the display name Discord rolled out in 2023 —
// preferred over Username for the seat label when non-empty, so
// users see "Alice" rather than "alice.1234" when Discord shows
// them the former.
type User struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name,omitempty"`
	Avatar     string `json:"avatar,omitempty"`
}

// DisplayName returns the preferred seat label: global_name first,
// then username. Empty only if Discord returned a completely blank
// record, which shouldn't happen but is handled by the caller
// falling back to the manual-name form field.
func (u User) DisplayName() string {
	if u.GlobalName != "" {
		return u.GlobalName
	}
	return u.Username
}

// FetchUser calls /users/@me with the access token from
// ExchangeCode and returns the parsed identity. Any network /
// status error surfaces as-is so the caller can log the full
// reason (usually rate-limited or token revoked).
func (c Config) FetchUser(ctx context.Context, client *http.Client, accessToken string) (User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userEndpoint, nil)
	if err != nil {
		return User{}, fmt.Errorf("discord: build user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("discord: user request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("discord: /users/@me: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
	if err != nil {
		return User{}, fmt.Errorf("discord: read user body: %w", err)
	}
	var u User
	if err := json.Unmarshal(body, &u); err != nil {
		return User{}, fmt.Errorf("discord: parse user body: %w", err)
	}
	if u.ID == "" {
		return User{}, errors.New("discord: /users/@me returned empty id")
	}
	return u, nil
}
