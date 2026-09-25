package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

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
	// ErrGameNotFound is returned by GetGame and ArchiveGame on a 404
	// — either the ID is wrong or the game was deleted (not merely
	// archived: GET /games/{id} and POST /games/{id}/archive both
	// still find an archived game, since neither is the listing
	// endpoint that filters them out).
	ErrGameNotFound = errors.New("game not found")
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
	defer func() { _ = resp.Body.Close() }()

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
// invite token the bot posts back to Discord. hostDiscordID, when
// non-empty, names the table host (ADR 0075 §2.1): the server binds
// hosting to that Discord user's seat once they claim one.
func (c *ServerClient) CreateGame(ctx context.Context, name, hostDiscordID string) (lobby.GameMeta, error) {
	req := map[string]string{"name": name}
	if hostDiscordID != "" {
		req["host_discord_id"] = hostDiscordID
	}
	body, err := json.Marshal(req)
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
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = resp.Body.Close() }()

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

// GetGame calls GET /games/{id} with the cached admin session
// (re-logging in once on a 401). Unlike ListGames, this reaches
// archived games too — Lobby.Get does not filter them the way
// Lobby.List does — which is what lets /c2-end tell "already
// archived" apart from "no such game" before it asks for
// confirmation.
func (c *ServerClient) GetGame(ctx context.Context, id uuid.UUID) (lobby.GameMeta, error) {
	resp, err := c.doAuthorized(ctx, func(token string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/games/"+id.String(), nil)
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
		return lobby.GameMeta{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return lobby.GameMeta{}, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusNotFound {
		return lobby.GameMeta{}, ErrGameNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return lobby.GameMeta{}, statusErr(resp)
	}

	var meta lobby.GameMeta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return lobby.GameMeta{}, fmt.Errorf("decode game: %w", err)
	}
	return meta, nil
}

// IsCreator calls GET /games/{id}/creator?discord_id=<discordID> with
// the cached admin session (re-logging in once on a 401) — the
// server-side half of #1098's /c2-end host check: does discordID
// match the Discord user who created this game. The server never
// says who the creator actually is; only whether this one snowflake
// matches. ErrGameNotFound on a 404, same as GetGame and ArchiveGame.
func (c *ServerClient) IsCreator(ctx context.Context, id uuid.UUID, discordID string) (bool, error) {
	resp, err := c.doAuthorized(ctx, func(token string) (*http.Response, error) {
		u := c.baseURL + "/games/" + id.String() + "/creator?discord_id=" + url.QueryEscape(discordID)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return false, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, ErrGameNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return false, statusErr(resp)
	}

	var body struct {
		IsCreator bool `json:"is_creator"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, fmt.Errorf("decode creator check: %w", err)
	}
	return body.IsCreator, nil
}

// ArchiveGame calls POST /games/{id}/archive with the cached admin
// session (re-logging in once on a 401). The route is idempotent
// server-side (lobby.Lobby.SetArchived): archiving an
// already-archived game still returns 200 with the same
// archived_at, rather than an error.
func (c *ServerClient) ArchiveGame(ctx context.Context, id uuid.UUID) (lobby.GameMeta, error) {
	resp, err := c.doAuthorized(ctx, func(token string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/games/"+id.String()+"/archive", nil)
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
		return lobby.GameMeta{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return lobby.GameMeta{}, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusNotFound {
		return lobby.GameMeta{}, ErrGameNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return lobby.GameMeta{}, statusErr(resp)
	}

	var meta lobby.GameMeta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return lobby.GameMeta{}, fmt.Errorf("decode game: %w", err)
	}
	return meta, nil
}

// --- Deck coverage / deck requests (ADR 0095 §4, #1631) ---
//
// These types mirror the wire shapes documented in docs/lobby.md's
// "POST /deck-coverage" and "POST /deck-requests" sections rather
// than importing server/internal/deckcoverage — that package (PR 1
// of ADR 0095) may not exist in every tree this package builds
// against, and the bot only ever needs the JSON contract, never the
// server's internal Go types.

// DeckCoverageBucket is one of the five buckets POST /deck-coverage
// sorts a decklist's cards into.
type DeckCoverageBucket string

// The five buckets, in the order POST /deck-coverage sorts cards:
// most actionable first.
const (
	BucketManual     DeckCoverageBucket = "manual"
	BucketUnreviewed DeckCoverageBucket = "unreviewed"
	BucketCaveats    DeckCoverageBucket = "caveats"
	BucketAutomated  DeckCoverageBucket = "automated"
	BucketNoEffect   DeckCoverageBucket = "no_effect"
)

// DeckViolation is one entry of deck.Validate's informational output,
// carried on both a successful report and a fetch-error response.
type DeckViolation struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Card    string `json:"card,omitempty"`
}

// DeckCoverageCard is one card's line in a coverage report.
type DeckCoverageCard struct {
	Name     string             `json:"name"`
	OracleID string             `json:"oracle_id"`
	Count    int                `json:"count"`
	Bucket   DeckCoverageBucket `json:"bucket"`
	Caveats  []string           `json:"caveats,omitempty"`
}

// DeckCoverageReport is the body of a successful POST /deck-coverage
// response, and the `report` field of a POST /deck-requests response.
type DeckCoverageReport struct {
	DeckName   string                     `json:"deck_name"`
	Source     string                     `json:"source"`
	SourceURL  string                     `json:"source_url,omitempty"`
	DeckKey    string                     `json:"deck_key,omitempty"`
	Commanders []string                   `json:"commanders"`
	Counts     map[DeckCoverageBucket]int `json:"counts"`
	Cards      []DeckCoverageCard         `json:"cards"`
	Unknown    []string                   `json:"unknown"`
	Violations []DeckViolation            `json:"violations"`
}

// DeckAPIError is returned by DeckCoverage and DeckRequest for any
// non-2xx response the two routes' error tables describe (docs/lobby.md):
// a fetch error with a code and violations, or the uniform {error}
// shape a bad body / 503 / 502 uses. Message is always the
// player-readable sentence the server sent.
type DeckAPIError struct {
	StatusCode int
	Message    string
	Code       string
	Violations []DeckViolation
}

func (e *DeckAPIError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("server returned %d", e.StatusCode)
}

// DeckCoverage calls POST /deck-coverage with the cached admin
// session (re-logging in once on a 401) — the bot's own, larger rate
// bucket rather than the public per-IP one (ADR 0095 §4). link must
// be a Moxfield or Archidekt deck URL.
func (c *ServerClient) DeckCoverage(ctx context.Context, link string) (DeckCoverageReport, error) {
	body, err := json.Marshal(map[string]string{"url": link})
	if err != nil {
		return DeckCoverageReport{}, err
	}
	resp, err := c.doAuthorized(ctx, func(token string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/deck-coverage", bytes.NewReader(body))
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
		return DeckCoverageReport{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return DeckCoverageReport{}, ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return DeckCoverageReport{}, decodeDeckAPIError(resp)
	}

	var report DeckCoverageReport
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return DeckCoverageReport{}, fmt.Errorf("decode deck-coverage response: %w", err)
	}
	return report, nil
}

// DeckRequester is the {discord_id, display_name} the bot sends on
// POST /deck-requests, naming the guild member it is asking for.
type DeckRequester struct {
	DiscordID   string `json:"discord_id"`
	DisplayName string `json:"display_name"`
}

// DeckRequestStatus is POST /deck-requests's `status` field.
type DeckRequestStatus string

const (
	DeckRequestFiled        DeckRequestStatus = "filed"
	DeckRequestJoined       DeckRequestStatus = "joined"
	DeckRequestNothingToAdd DeckRequestStatus = "nothing_to_add"
	DeckRequestRateLimited  DeckRequestStatus = "rate_limited"
)

// DeckRequestResult is a successful (2xx or 429-with-status) response
// from POST /deck-requests. Report is nil only when the server sent
// none (defensive; every documented status carries one).
type DeckRequestResult struct {
	Status           DeckRequestStatus   `json:"status"`
	IssueURL         string              `json:"issue_url,omitempty"`
	IssueNumber      int                 `json:"issue_number,omitempty"`
	RetryAfter       int                 `json:"retry_after,omitempty"`
	AlreadyRequested bool                `json:"already_requested,omitempty"`
	Report           *DeckCoverageReport `json:"report,omitempty"`
}

// DeckRequest calls POST /deck-requests with the cached admin session
// (re-logging in once on a 401), naming requester as the guild member
// the request is filed on behalf of. link must be a deck URL — the
// route refuses pasted text.
//
// A 429 carrying {"status":"rate_limited", ...} is a normal result,
// not an error: it is one of the four documented outcomes. A 429
// carrying only {"error":...} (the IP limiter's shape, which this
// route should not reach given the admin session, but is handled
// defensively) surfaces as a DeckAPIError instead.
func (c *ServerClient) DeckRequest(ctx context.Context, link string, requester DeckRequester) (DeckRequestResult, error) {
	body, err := json.Marshal(struct {
		URL       string        `json:"url"`
		Requester DeckRequester `json:"requester"`
	}{URL: link, Requester: requester})
	if err != nil {
		return DeckRequestResult{}, err
	}
	resp, err := c.doAuthorized(ctx, func(token string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/deck-requests", bytes.NewReader(body))
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
		return DeckRequestResult{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusCreated, http.StatusOK, http.StatusTooManyRequests:
		var result DeckRequestResult
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return DeckRequestResult{}, fmt.Errorf("decode deck-request response: %w", err)
		}
		if result.Status == "" {
			// A 429 with no status field is the generic IP-limiter
			// shape ({"error": "too many requests"}), not a
			// rate_limited deck-request result.
			return DeckRequestResult{}, &DeckAPIError{StatusCode: resp.StatusCode, Message: "too many requests"}
		}
		return result, nil
	case http.StatusUnauthorized:
		return DeckRequestResult{}, ErrUnauthorized
	default:
		return DeckRequestResult{}, decodeDeckAPIError(resp)
	}
}

// decodeDeckAPIError reads a non-2xx deck-coverage/deck-requests body
// into a *DeckAPIError. Both routes use the same two error shapes: a
// fetch error ({error, code, violations}) and the uniform {error}
// shape everything else uses.
func decodeDeckAPIError(resp *http.Response) error {
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var body struct {
		Error      string          `json:"error"`
		Code       string          `json:"code"`
		Violations []DeckViolation `json:"violations"`
	}
	_ = json.Unmarshal(b, &body)
	msg := body.Error
	if msg == "" {
		msg = strings.TrimSpace(string(b))
	}
	return &DeckAPIError{StatusCode: resp.StatusCode, Message: msg, Code: body.Code, Violations: body.Violations}
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
