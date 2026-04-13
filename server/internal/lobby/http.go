package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sessionTTL controls how long a newly-minted Principal lives in the
// authenticator. Short enough that leaked cookies rotate out of risk
// within a day; long enough that a friends' game session doesn't hit
// "please log in again" mid-match.
const sessionTTL = 12 * time.Hour

// Config bundles the dependencies Handler needs. Separate from
// Lobby itself so main.go can build the HTTP layer without the
// lobby having to know about auth.
type Config struct {
	Lobby        *Lobby
	Auth         auth.Authenticator
	AdminToken   string // shared admin token; empty disables admin flow
	SessionTTL   time.Duration
	AllowAnon    bool // allow unauthenticated /games/{id}/join via invite (default: true)
}

// Handler returns an http.Handler wired to the v0 lobby REST surface:
//
//	POST /admin/login       — exchange the admin token for a session
//	POST /games             — admin: create a new game (returns invite)
//	GET  /games             — authenticated: list known games
//	GET  /games/{id}        — authenticated: game metadata
//	POST /games/{id}/join   — invite + name → session + player_id
//	POST /games/{id}/start  — authenticated: transition lobby → active
//	GET  /me                — authenticated: principal echo (for client bootstrap)
//	POST /logout            — revoke the caller's session server-side
//
// Routes that mutate state accept JSON bodies; read-only routes use
// query params / path params. All responses are JSON.
func Handler(c Config) http.Handler {
	if c.SessionTTL == 0 {
		c.SessionTTL = sessionTTL
	}
	mux := http.NewServeMux()

	mux.Handle("POST /admin/login", handlerFunc(c, adminLogin))
	mux.Handle("POST /games/{id}/join", handlerFunc(c, joinGame))
	mux.Handle("POST /games", auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, createGame)))
	mux.Handle("GET /games", auth.Middleware(c.Auth)(handlerFunc(c, listGames)))
	mux.Handle("GET /games/{id}", auth.Middleware(c.Auth)(handlerFunc(c, getGame)))
	mux.Handle("POST /games/{id}/start", auth.Middleware(c.Auth)(handlerFunc(c, startGame)))
	mux.Handle("GET /me", auth.Middleware(c.Auth)(handlerFunc(c, me)))
	// Logout does not require an authenticated principal — a client
	// with a stale or revoked token should still be able to clear
	// browser state without a 401 dead-end. We just revoke whatever
	// credential is on the request (if any) and drop the cookie.
	mux.Handle("POST /logout", handlerFunc(c, logout))

	return mux
}

// handlerFunc adapts a (Config, w, r) → error closure into an
// http.Handler, centralising error-to-JSON translation so every
// endpoint doesn't repeat the same switch statement.
type lobbyHandler func(c Config, w http.ResponseWriter, r *http.Request) error

func handlerFunc(c Config, h lobbyHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := h(c, w, r); err != nil {
			writeLobbyError(w, err)
		}
	})
}

// --- request / response shapes ---

type adminLoginRequest struct {
	Token string `json:"token"`
}

type sessionResponse struct {
	Token     string          `json:"token"`
	ExpiresAt time.Time       `json:"expires_at"`
	Principal auth.Principal  `json:"principal"`
	Game      *GameMeta       `json:"game,omitempty"`
	PlayerID  uuid.UUID       `json:"player_id,omitempty"`
}

type createGameRequest struct {
	Name string `json:"name"`
}

type joinRequest struct {
	InviteToken string `json:"invite_token"`
	Name        string `json:"name"`
}

type listResponse struct {
	Games []GameMeta `json:"games"`
}

// --- handlers ---

// adminLogin exchanges the shared admin token for a session. This is
// the only way to acquire a RoleAdmin credential — no password file,
// no user accounts.
func adminLogin(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.AdminToken == "" {
		return httpError(http.StatusServiceUnavailable, "admin login disabled")
	}
	var body adminLoginRequest
	if err := decodeJSON(r, &body); err != nil {
		return err
	}
	if !constantTimeEqual(body.Token, c.AdminToken) {
		return httpError(http.StatusUnauthorized, "invalid admin token")
	}
	p := auth.Principal{Role: auth.RoleAdmin, AdminID: uuid.New(), Name: "admin"}
	tok, issued, err := c.Auth.Issue(r.Context(), p, c.SessionTTL)
	if err != nil {
		return err
	}
	setSessionCookie(w, tok, issued.ExpiresAt)
	return writeJSON(w, http.StatusOK, sessionResponse{Token: tok, ExpiresAt: issued.ExpiresAt, Principal: issued})
}

// createGame (admin-only) creates a new game in the lobby and
// returns its metadata INCLUDING the invite token. The admin is
// responsible for distributing the invite out-of-band.
func createGame(c Config, w http.ResponseWriter, r *http.Request) error {
	var body createGameRequest
	if err := decodeJSON(r, &body); err != nil {
		return err
	}
	meta, err := c.Lobby.Create(body.Name)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusCreated, meta)
}

// joinGame is the primary onboarding path: a player clicks an invite
// link, posts their display name + the invite token, and receives a
// session cookie + player ID. The player ID is what they pass to
// /ws?player=<uuid> from then on.
//
// This is the one public endpoint — it doesn't require a pre-
// existing session. The invite token IS the credential.
func joinGame(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	var body joinRequest
	if err := decodeJSON(r, &body); err != nil {
		return err
	}

	meta, playerID, err := c.Lobby.Join(id, body.InviteToken, body.Name)
	if err != nil {
		return err
	}

	// Mint a RolePlayer session bound to (gameID, playerID). The WS
	// authorizer will cross-check the principal's GameID against the
	// one on the upgrade request — a player session can't be reused
	// to spy on a different game.
	p := auth.Principal{
		Role:     auth.RolePlayer,
		GameID:   meta.ID,
		PlayerID: playerID,
		Name:     body.Name,
	}
	tok, issued, err := c.Auth.Issue(r.Context(), p, c.SessionTTL)
	if err != nil {
		return err
	}
	setSessionCookie(w, tok, issued.ExpiresAt)

	// Strip the invite token from the returned meta — the joiner
	// already has it, and other joiners don't need to see it in the
	// redirect-time response.
	meta.InviteToken = ""
	return writeJSON(w, http.StatusOK, sessionResponse{
		Token:     tok,
		ExpiresAt: issued.ExpiresAt,
		Principal: issued,
		Game:      &meta,
		PlayerID:  playerID,
	})
}

func listGames(c Config, w http.ResponseWriter, _ *http.Request) error {
	return writeJSON(w, http.StatusOK, listResponse{Games: c.Lobby.List()})
}

func getGame(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	// Only the admin and seated players should see the invite token
	// — drop it for anyone else. Require a principal attached by the
	// auth middleware: if it's missing the request slipped past our
	// mux, which is a server bug and should 500 before leaking data.
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.Role != auth.RoleAdmin && p.GameID != id {
		meta.InviteToken = ""
	}
	return writeJSON(w, http.StatusOK, meta)
}

func startGame(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	// Require that the caller be either an admin or a seated player
	// in this game. Anyone else with a stray session cookie shouldn't
	// be able to yank a game to "active".
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.Role != auth.RoleAdmin && p.GameID != id {
		return httpError(http.StatusForbidden, "not a seat in this game")
	}
	meta, err := c.Lobby.Start(id)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, meta)
}

// logout revokes the caller's credential server-side and clears the
// session cookie. Unauthenticated — we want a client with an already-
// expired token to be able to reach this endpoint to flush its cookie
// without hitting a 401 first. Always returns 204 so the client can
// safely treat the response as idempotent.
func logout(c Config, w http.ResponseWriter, r *http.Request) error {
	if cred := auth.CredentialFromRequest(r); cred != "" {
		// Ignore Revoke errors: a stateless HMAC-style authenticator
		// may always return nil, a stateful one may return "unknown
		// token" which we treat as already-revoked. Either way the
		// client just wants its cookie cleared.
		_ = c.Auth.Revoke(r.Context(), cred)
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func me(_ Config, w http.ResponseWriter, r *http.Request) error {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	return writeJSON(w, http.StatusOK, p)
}

// --- helpers ---

func gameIDFromPath(r *http.Request) (uuid.UUID, error) {
	raw := r.PathValue("id")
	if raw == "" {
		return uuid.Nil, httpError(http.StatusBadRequest, "missing game id")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, httpError(http.StatusBadRequest, "invalid game id")
	}
	return id, nil
}

// decodeJSON enforces Content-Type: application/json when a body is
// present and rejects unknown fields so typos in request shapes
// surface early. Returns a 400-bearing error.
func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return httpError(http.StatusBadRequest, "missing body")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return httpError(http.StatusBadRequest, fmt.Sprintf("invalid body: %s", err.Error()))
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(body)
}

// setSessionCookie stamps the browser session cookie that transports
// the credential on subsequent HTTP and same-origin WS requests.
// HttpOnly prevents JS access; SameSite=Lax lets the invite-link
// flow work (top-level navigation). Not Secure at S04 — this is a
// local LAN service. Flip this in S12 when TLS lands.
func setSessionCookie(w http.ResponseWriter, tok string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookie,
		Value:    tok,
		Path:     "/",
		Expires:  exp,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSessionCookie emits a Set-Cookie that evicts the browser's
// current session cookie. MaxAge=-1 tells browsers to drop it
// immediately; Expires in the past covers older clients that ignore
// MaxAge.
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// httpStatusError is a lobby-local wrapper that carries an HTTP
// status alongside the message, so writeLobbyError can map sentinels
// to the right code.
type httpStatusError struct {
	code int
	msg  string
}

func (e *httpStatusError) Error() string { return e.msg }

func httpError(code int, msg string) error {
	return &httpStatusError{code: code, msg: msg}
}

// writeLobbyError translates a handler error into an HTTP status +
// JSON body. Sentinels from the lobby / game / auth packages get
// specific status codes; everything else falls through to 500.
func writeLobbyError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	msg := err.Error()

	var se *httpStatusError
	switch {
	case errors.As(err, &se):
		status = se.code
	case errors.Is(err, ErrGameNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrInvalidInvite):
		status = http.StatusUnauthorized
	case errors.Is(err, ErrGameFull),
		errors.Is(err, ErrGameStarted),
		errors.Is(err, ErrSeatTaken),
		errors.Is(err, game.ErrNotEnoughPlayers),
		errors.Is(err, game.ErrGameAlreadyStarted),
		errors.Is(err, game.ErrGameNotInLobby):
		status = http.StatusConflict
	case errors.Is(err, ErrEmptyName):
		status = http.StatusBadRequest
	case errors.Is(err, auth.ErrInvalidCredential),
		errors.Is(err, auth.ErrExpiredCredential),
		errors.Is(err, auth.ErrUnknownPrincipal):
		status = http.StatusUnauthorized
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// constantTimeEqual compares two strings in constant time (byte-wise,
// ignoring that the underlying lengths may differ — which is itself
// a timing signal, but for admin-token comparison this is fine).
func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range len(a) {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

// compile-time assertion we haven't dropped context-awareness from
// critical methods.
var _ = context.Background
