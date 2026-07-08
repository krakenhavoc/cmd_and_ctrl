package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
)

// Default request timeout. The loopback calls we make (admin
// login, create game, list games) should resolve in well under
// 100 ms on the VPS. 5 s is generous headroom for a GC pause or
// a cold-started server; the 3 s Discord interaction deadline
// forces us to keep this below the SDK's own budget.
const defaultRequestTimeout = 5 * time.Second

// Errors returned by ServerClient. Wrapped by the command
// handlers into user-visible messages; callers should use
// errors.Is to distinguish auth from generic failures.
var (
	ErrServerUnreachable = errors.New("game server is not reachable")
	ErrUnauthorized      = errors.New("bot is not authorized against the game server")
)

// ServerClient is the bot's view of the game server over
// loopback HTTP. It holds a cached admin session token and
// refreshes on demand when the server rejects it — the two-step
// pattern keeps the hot path (create / list) a single HTTP call
// when the token is still valid, and stops every slash command
// from minting a fresh 12h admin session that piles up server-side.
//
// Safe for concurrent use; the token mutex guards a single
// string swap.
type ServerClient struct {
	baseURL    string
	adminToken string
	http       *http.Client

	mu      sync.Mutex
	session string // cached admin session bearer; "" = not logged in
}

// NewServerClient builds a client rooted at baseURL. The
// adminToken is the same shared secret the game server carries
// as CMDCTRL_ADMIN_TOKEN.
func NewServerClient(baseURL, adminToken string) *ServerClient {
	return &ServerClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		adminToken: adminToken,
		http:       &http.Client{Timeout: defaultRequestTimeout},
	}
}

// WithHTTPClient injects a custom http.Client. Intended for
// tests that drive against an httptest.Server.
func (c *ServerClient) WithHTTPClient(h *http.Client) *ServerClient {
	c.http = h
	return c
}

// loginResponse mirrors the subset of lobby.sessionResponse the
// bot cares about. The full shape carries more (Principal, Game,
// PlayerID), but the bot only needs the bearer.
type loginResponse struct {
	Token string `json:"token"`
}

// listResponse mirrors lobby.listResponse. Re-declared here
// because that type is unexported in the lobby package.
type listResponse struct {
	Games []lobby.GameMeta `json:"games"`
}

// Login exchanges the admin shared-secret for a session bearer.
// Returns ErrUnauthorized on 401 so callers can surface a crisp
// "check CMDCTRL_ADMIN_TOKEN" error; other non-2xx results are
// wrapped with the status text.
func (c *ServerClient) Login(ctx context.Context) (string, error) {
	body, err := json.Marshal(map[string]string{"token": c.adminToken})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/admin/login", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrServerUnreachable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return "", statusErr(resp)
	}

	var lr loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return "", fmt.Errorf("decode login response: %w", err)
	}
	if lr.Token == "" {
		return "", errors.New("login response missing token")
	}
	return lr.Token, nil
}

// sessionToken returns the cached admin session bearer, logging in
// to mint one only when the cache is empty. Held under the mutex so
// concurrent slash commands share a single login round-trip.
func (c *ServerClient) sessionToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session != "" {
		return c.session, nil
	}
	tok, err := c.Login(ctx)
	if err != nil {
		return "", err
	}
	c.session = tok
	return tok, nil
}

// invalidateSession drops the cached bearer iff it still equals the
// one the server just rejected — a concurrent command may already
// have replaced it with a fresh login, which must not be discarded.
func (c *ServerClient) invalidateSession(stale string) {
	c.mu.Lock()
	if c.session == stale {
		c.session = ""
	}
	c.mu.Unlock()
}

// doAuthorized runs `do` with a session token, re-logging in and
// retrying exactly once when the server 401s the cached token
// (expired TTL, server restart). `do` must build a fresh
// *http.Request per call — request bodies are single-use. A 401 on
// the retry surfaces to the caller as a normal response.
func (c *ServerClient) doAuthorized(ctx context.Context, do func(token string) (*http.Response, error)) (*http.Response, error) {
	tok, err := c.sessionToken(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := do(tok)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}
	_ = resp.Body.Close()
	c.invalidateSession(tok)
	tok, err = c.sessionToken(ctx)
	if err != nil {
		return nil, err
	}
	return do(tok)
}

// CreateGame calls POST /games with the cached admin session
// (re-logging in once on a 401). The returned GameMeta carries the
// invite token the bot posts back to Discord.
func (c *ServerClient) CreateGame(ctx context.Context, name string) (lobby.GameMeta, error) {
	body, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		return lobby.GameMeta{}, err
	}
	resp, err := c.doAuthorized(ctx, func(token string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/games", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrServerUnreachable, err)
		}
		return resp, nil
	})
	if err != nil {
		return lobby.GameMeta{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return lobby.GameMeta{}, ErrUnauthorized
	}
	if resp.StatusCode != http.StatusCreated {
		return lobby.GameMeta{}, statusErr(resp)
	}

	var meta lobby.GameMeta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return lobby.GameMeta{}, fmt.Errorf("decode game: %w", err)
	}
	if meta.InviteToken == "" {
		return lobby.GameMeta{}, errors.New("create-game response missing invite_token")
	}
	return meta, nil
}

// ListGames calls GET /games with the cached admin session
// (re-logging in once on a 401). The server already strips invite
// tokens from list responses (see lobby.Lobby.List), so the
// slice is safe to forward directly to a Discord channel.
func (c *ServerClient) ListGames(ctx context.Context) ([]lobby.GameMeta, error) {
	resp, err := c.doAuthorized(ctx, func(token string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/games", nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrServerUnreachable, err)
		}
		return resp, nil
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return nil, statusErr(resp)
	}

	var lr listResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("decode list: %w", err)
	}
	return lr.Games, nil
}

// statusErr builds an informative error for an unexpected
// status code. Reads at most 4 KiB of body so a runaway
// response doesn't blow the bot's memory.
func statusErr(resp *http.Response) error {
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	snippet := strings.TrimSpace(string(b))
	if snippet != "" {
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, snippet)
	}
	return fmt.Errorf("server returned %d", resp.StatusCode)
}
