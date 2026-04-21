package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/ratelimit"
)

// sessionTTL controls how long a newly-minted Principal lives in the
// authenticator. Short enough that leaked cookies rotate out of risk
// within a day; long enough that a friends' game session doesn't hit
// "please log in again" mid-match.
const sessionTTL = 12 * time.Hour

// maxDeckBodyBytes caps the payload accepted by POST /games/{id}/decks.
// A 100-card Moxfield JSON export is typically well under 200 KiB; 2 MiB
// is generous headroom and still prohibitive as a DoS primitive.
const maxDeckBodyBytes = 2 * 1024 * 1024

// Config bundles the dependencies Handler needs. Separate from
// Lobby itself so main.go can build the HTTP layer without the
// lobby having to know about auth.
type Config struct {
	Lobby      *Lobby
	Auth       auth.Authenticator
	AdminToken string // shared admin token; empty disables admin flow
	SessionTTL time.Duration
	AllowAnon  bool // allow unauthenticated /games/{id}/join via invite (default: true)
	// Cards is the Scryfall index used by the deck-upload endpoint.
	// When nil, POST /games/{id}/decks returns 503 so a fresh
	// deployment (no Scryfall dump yet) surfaces a clear "run
	// scryfall-refresh.sh" error rather than a cryptic unknown-card
	// list.
	Cards *cards.Index
	// Evictor optionally closes any WS clients bound to a game when
	// the game is deleted. When nil, DELETE still drops the game
	// from the lobby + room manager but existing sockets linger
	// until their next action hits "game no longer available".
	Evictor GameEvictor
	// DeckHTTPClient is the outbound client used by the S06.5 URL-
	// based deck import (`format: "url"`). Nil falls back to a
	// sensible default with a 10 s timeout and the project's User-Agent.
	// Tests inject an httptest.Server-backed client here so the
	// upload path never touches the live Moxfield / Archidekt APIs.
	DeckHTTPClient *http.Client

	// Discord carries the S12.5 OAuth config. When Enabled() is
	// false the /auth/discord/* routes are still registered but
	// return 503, and /auth/discord/config reports enabled: false
	// so the client hides the sign-in button. Unset in tests that
	// don't exercise the OAuth surface.
	Discord discord.Config
	// DiscordStateStore holds in-flight OAuth rounds. Nil falls
	// back to a freshly-constructed store on first use so tests
	// that don't set it still work; production wires a single
	// shared store at boot so concurrent OAuths don't each build
	// their own map.
	DiscordStateStore *discord.StateStore
	// DiscordHTTPClient is injected for tests that stub Discord's
	// token + /users/@me endpoints via httptest. Nil falls back to
	// http.DefaultClient.
	DiscordHTTPClient *http.Client

	// DiscordAvatars is the S12.5 avatar cache. Nil means
	// /avatars/* returns 503; production wires a cache rooted at
	// $CMDCTRL_DATA_DIR/avatars so disk-cached images persist
	// across restarts. The client falls back to an initials
	// placeholder when the endpoint 503s, so an unconfigured
	// deploy still renders a usable board.
	DiscordAvatars *discord.AvatarCache
}

// GameEvictor is the subset of *ws.Hub that the lobby needs to close
// connections for a deleted game. Interface rather than concrete type
// so tests can omit it without standing up a hub.
type GameEvictor interface {
	EvictGame(gameID uuid.UUID) int
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

	// Rate-limit the two endpoints that accept untrusted credentials:
	// /admin/login brute-forces the shared admin token, /join
	// brute-forces invite tokens. 1 req/s with a 5-token burst per
	// client IP is lenient for real humans, prohibitive for scripts.
	limit := ratelimit.New(1, 5)
	// Deck upload has a separate, more permissive bucket: legitimate
	// players may re-upload several times while iterating, but we still
	// want a ceiling on how fast a single IP can stream megabyte-sized
	// Moxfield blobs at the parser.
	deckLimit := ratelimit.New(2, 10)
	mux.Handle("POST /admin/login", limit.Middleware(handlerFunc(c, adminLogin)))
	mux.Handle("POST /games/{id}/join", limit.Middleware(handlerFunc(c, joinGame)))
	mux.Handle("POST /games/{id}/spectate", limit.Middleware(handlerFunc(c, spectateGame)))

	// Discord OAuth (S12.5). All three are unauthenticated — they
	// either run before any session exists or carry their own
	// CSRF protection via the state token. /start and /callback
	// are rate-limited alongside admin-login + join because state
	// generation involves crypto/rand and the upstream calls are
	// the most expensive thing we forward to Discord.
	mux.Handle("GET /auth/discord/config", handlerFunc(c, discordConfig))
	mux.Handle("GET /auth/discord/start", limit.Middleware(handlerFunc(c, discordStart)))
	mux.Handle("GET /auth/discord/callback", limit.Middleware(handlerFunc(c, discordCallback)))
	// Discord avatar cache. Unauthenticated — the avatar hashes
	// are effectively public (anyone in a shared guild already
	// sees them via Discord's CDN) and a cached image tells the
	// requester nothing they couldn't learn by watching the
	// game's seat list. An authenticated gate would mean every
	// <img> in the board has to carry a token.
	mux.Handle("GET /avatars/{id}/{hash}", handlerFunc(c, discordAvatar))
	mux.Handle("POST /games", auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, createGame)))
	mux.Handle("DELETE /games/{id}", auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, deleteGame)))
	mux.Handle("GET /games", auth.Middleware(c.Auth)(handlerFunc(c, listGames)))
	mux.Handle("GET /games/{id}", auth.Middleware(c.Auth)(handlerFunc(c, getGame)))
	mux.Handle("POST /games/{id}/start", auth.Middleware(c.Auth)(handlerFunc(c, startGame)))
	mux.Handle("GET /games/{id}/replay", auth.Middleware(c.Auth)(handlerFunc(c, downloadReplay)))
	mux.Handle("POST /games/{id}/decks", deckLimit.Middleware(auth.Middleware(c.Auth)(handlerFunc(c, uploadDeck))))
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
	Token     string         `json:"token"`
	ExpiresAt time.Time      `json:"expires_at"`
	Principal auth.Principal `json:"principal"`
	Game      *GameMeta      `json:"game,omitempty"`
	PlayerID  uuid.UUID      `json:"player_id,omitempty"`
}

type createGameRequest struct {
	Name string `json:"name"`
}

type joinRequest struct {
	InviteToken string `json:"invite_token"`
	Name        string `json:"name"`
}

// spectateRequest is the body of POST /games/{id}/spectate. The
// invite token is the per-game spectator invite from GameMeta. Name
// is purely cosmetic (chat author label, future presence indicator).
type spectateRequest struct {
	InviteToken string `json:"invite_token"`
	Name        string `json:"name,omitempty"`
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

	// Strip both invite tokens from the returned meta — the joiner
	// already has the player invite they used to get here, and the
	// spectator invite is admin-only data that shouldn't leak to a
	// freshly-seated player by default.
	meta.InviteToken = ""
	meta.SpectatorInvite = ""
	return writeJSON(w, http.StatusOK, sessionResponse{
		Token:     tok,
		ExpiresAt: issued.ExpiresAt,
		Principal: issued,
		Game:      &meta,
		PlayerID:  playerID,
	})
}

// spectateGame is the spectator counterpart to joinGame: a viewer
// posts the per-game spectator invite and receives a RoleSpectator
// session bound to the game (no PlayerID). They can then open
// /ws?game=<id>&token=<tok> with no `?player=` and watch.
//
// No deck or seat is allocated. Spectators may join in any game
// state (lobby / active / ended).
func spectateGame(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	var body spectateRequest
	if err := decodeJSON(r, &body); err != nil {
		return err
	}

	meta, err := c.Lobby.Spectate(id, body.InviteToken)
	if err != nil {
		return err
	}

	p := auth.Principal{
		Role:   auth.RoleSpectator,
		GameID: meta.ID,
		Name:   trimToLimit(body.Name, 40),
	}
	tok, issued, err := c.Auth.Issue(r.Context(), p, c.SessionTTL)
	if err != nil {
		return err
	}
	setSessionCookie(w, tok, issued.ExpiresAt)

	// Spectators don't see either invite — neither the player nor
	// the spectator one. They have what they need.
	meta.InviteToken = ""
	meta.SpectatorInvite = ""
	return writeJSON(w, http.StatusOK, sessionResponse{
		Token:     tok,
		ExpiresAt: issued.ExpiresAt,
		Principal: issued,
		Game:      &meta,
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
		meta.SpectatorInvite = ""
	}
	// Spectators specifically never see the player invite (would let
	// them claim a seat) and shouldn't see the spectator one either —
	// their session is already proof they have it.
	if p.Role == auth.RoleSpectator {
		meta.InviteToken = ""
		meta.SpectatorInvite = ""
	}
	return writeJSON(w, http.StatusOK, meta)
}

// deleteGame (admin-only) removes a game from the lobby and from the
// underlying RoomManager, and evicts any WebSocket clients currently
// bound to it. Idempotent only in the sense that a second DELETE
// returns 404 rather than an error shape — the game is gone.
func deleteGame(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	if err := c.Lobby.Delete(id); err != nil {
		return err
	}
	if c.Evictor != nil {
		c.Evictor.EvictGame(id)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// downloadReplay streams the per-game JSONL replay log back to the
// caller. Each line is one protocol.SnapshotPayload (the same shape
// the WS layer broadcasts), in the order Apply produced them. A
// downstream consumer can deserialize line-by-line to reconstruct
// the full game timeline.
//
// Authorization: admin or any seat (player or spectator) in this
// game. The replay holds opponent hand contents in pre-filter form,
// so it MUST NOT be served to unauthenticated callers — but a player
// who was at the table has already seen these snapshots filtered for
// their seat, and an admin / spectator who watched live likewise.
// Sandbox-friendly default; if a deployment wants stricter
// "admin-only replays", flip the role check below.
func downloadReplay(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.Role != auth.RoleAdmin && p.GameID != id {
		return httpError(http.StatusForbidden, "not a seat in this game")
	}
	room := c.Lobby.RoomOf(id)
	if room == nil {
		return httpError(http.StatusNotFound, "game not found")
	}
	path := room.ReplayPath()
	if path == "" {
		return httpError(http.StatusServiceUnavailable, "replay log disabled (no dump dir)")
	}
	// Surface the file's existence cleanly: a brand-new game with no
	// Apply yet has no replay file. Returning 204 keeps the contract
	// "the route exists; there's just nothing to stream" without
	// 404-confusing the client.
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNoContent)
			return nil
		}
		return httpError(http.StatusInternalServerError, "stat replay: "+err.Error())
	}
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set(
		"Content-Disposition",
		`attachment; filename="`+id.String()+`.jsonl"`,
	)
	http.ServeContent(w, r, id.String()+".jsonl", info.ModTime(), mustOpen(path))
	return nil
}

// mustOpen returns an *os.File for ServeContent. ServeContent needs
// an io.ReadSeeker; os.File satisfies it. The defer-close lives in
// http.ServeContent's lifetime via the response writer's flush — we
// can't defer here because ServeContent reads from the file after
// this function returns.
func mustOpen(path string) *os.File {
	f, _ := os.Open(path)
	return f
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

// uploadDeckRequest is the request shape for POST /games/{id}/decks.
// Exactly one of Text / Moxfield must be set; Format is optional but
// helps the handler pick a parser when the caller can't match the
// file extension to a format string. The server echoes back a
// parsed + validated summary on success.
type uploadDeckRequest struct {
	// Format is one of "text", "moxfield", or empty (auto-detect
	// from the first non-whitespace byte: '{' → moxfield, else text).
	Format string `json:"format,omitempty"`
	// Source is the raw decklist payload. For "text" format, the
	// plain-text decklist. For "moxfield", the JSON export bytes.
	Source string `json:"source"`
	// PlayerID is the seat this deck is for. A RolePlayer caller can
	// only set their own deck — we cross-check against the principal
	// below.
	PlayerID uuid.UUID `json:"player_id"`
}

// uploadDeckResponse describes the accepted deck. Mirrors the lobby
// seat update the caller will see via GET /games/{id}, but returned
// inline so the client doesn't have to refetch.
type uploadDeckResponse struct {
	Game       GameMeta `json:"game"`
	DeckName   string   `json:"deck_name"`
	CardCount  int      `json:"card_count"`
	Commanders []string `json:"commanders"`
	// Warnings is a non-fatal violation list (e.g. sideboard ignored).
	// The accepted deck is already installed when warnings is
	// non-empty; treat it as advisory.
	Warnings []deck.Violation `json:"warnings,omitempty"`
}

// uploadDeck handles POST /games/{id}/decks. Accepts either plain-
// text or Moxfield JSON, resolves each card against the Scryfall
// index, validates the result against Commander rules, and — on
// success — replaces the seat's library + command zone on the
// authoritative game.
//
// Authorization:
//   - RolePlayer sessions may only set their OWN deck (p.PlayerID
//     must equal body.PlayerID and p.GameID must match the path id).
//   - RoleAdmin may set any seat's deck (useful for debugging and
//     for the rare "uploaded the wrong file" case).
//
// On validation failure the endpoint returns 422 with the full
// violation list so the client can highlight every offending card
// at once.
func uploadDeck(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.Cards == nil || c.Cards.Count() == 0 {
		return httpError(http.StatusServiceUnavailable, "card index not loaded; run scripts/scryfall-refresh.sh")
	}
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}

	// Cap the request body before decoding so a hostile client can't
	// stream megabytes at the parser. 2 MiB is roughly 10x the size of
	// a generous Moxfield JSON export.
	if r.Body == nil {
		return httpError(http.StatusBadRequest, "missing body")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxDeckBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var body uploadDeckRequest
	if err := dec.Decode(&body); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return httpError(http.StatusRequestEntityTooLarge, fmt.Sprintf("deck source exceeds %d-byte limit", maxDeckBodyBytes))
		}
		return httpError(http.StatusBadRequest, fmt.Sprintf("invalid body: %s", err.Error()))
	}
	if strings.TrimSpace(body.Source) == "" {
		return httpError(http.StatusBadRequest, "source is required")
	}
	if body.PlayerID == uuid.Nil {
		return httpError(http.StatusBadRequest, "player_id is required")
	}

	// Role enforcement: players must match their own principal.
	if p.Role == auth.RolePlayer {
		if p.GameID != id {
			return httpError(http.StatusForbidden, "session is not for this game")
		}
		if p.PlayerID != body.PlayerID {
			return httpError(http.StatusForbidden, "players may only upload their own deck")
		}
	}

	// Parse: auto-detect if Format is empty. URLs are detected first
	// since they're unambiguous ("http://" or "https://" prefix);
	// JSON next (leading `{`); everything else is plain text.
	format := body.Format
	source := strings.TrimLeft(body.Source, " \t\r\n")
	if format == "" {
		switch {
		case strings.HasPrefix(source, "http://"), strings.HasPrefix(source, "https://"):
			format = "url"
		case strings.HasPrefix(source, "{"):
			format = "moxfield"
		default:
			format = "text"
		}
	}

	var (
		deckName string
		entries  []deck.Entry
		perr     error
	)
	switch format {
	case "text":
		entries, perr = deck.ParseText(body.Source)
	case "moxfield":
		deckName, entries, perr = deck.ParseMoxfield([]byte(body.Source))
	case "url":
		client := c.DeckHTTPClient
		if client == nil {
			client = deck.DefaultClient()
		}
		deckName, entries, perr = deck.FetchFromURL(r.Context(), client, strings.TrimSpace(body.Source))
		if perr != nil {
			// Fetcher failures (unknown host, private deck, upstream
			// down, etc.) surface via FetchViolation as typed 422
			// entries so the client renders them the same way it
			// renders validation violations.
			if v, ok := deck.FetchViolation(strings.TrimSpace(body.Source), perr); ok {
				return writeDeckViolations(w, perr.Error(), []deck.Violation{v}, nil)
			}
			// Unknown-mechanic inside the fetched payload (e.g. a
			// Moxfield deck with a companion slot) bubbles through
			// here; fall into the existing Resolve-failure path
			// below by leaving perr set.
		}
	default:
		return httpError(http.StatusBadRequest, fmt.Sprintf("unknown deck format %q", format))
	}
	if perr != nil {
		return httpError(http.StatusBadRequest, perr.Error())
	}

	list, err := deck.Resolve(c.Cards, deckName, entries)
	if err != nil {
		// Resolve failures that carry a per-card Violation list are
		// surfaced with the same 422 `{"error", "violations"}` shape
		// the validator uses, so the client has one schema to handle.
		var uce *deck.UnknownCardError
		if errors.As(err, &uce) {
			return writeDeckViolations(w, err.Error(), uce.Violations(), nil)
		}
		var ume *deck.UnsupportedMechanicError
		if errors.As(err, &ume) {
			return writeDeckViolations(w, err.Error(), ume.Violations(), nil)
		}
		return httpError(http.StatusBadRequest, err.Error())
	}

	// Validate. Sideboard-only warnings are treated as non-fatal —
	// we strip them from the violation list and pass the rest
	// through. Everything else means the deck cannot be installed.
	//
	// Dev bypass (S14): when CMDCTRL_DEV_SKIP_DECK_VALIDATION is set
	// to a non-empty value, skip validation entirely. Intended for
	// manual-testing the card-effect catalog with a small throwaway
	// deck (~5 cards + commander) — a full 100-card deck is
	// tedious when all you want to do is cast Lightning Bolt. The
	// env var is read per-request so flipping it on/off on a live
	// server doesn't require a restart. Any non-empty value counts
	// as "on" — defer to shell truthiness. NEVER set this in
	// production.
	var warnings []deck.Violation
	var fatal []deck.Violation
	if os.Getenv("CMDCTRL_DEV_SKIP_DECK_VALIDATION") != "" {
		warnings = append(warnings, deck.Violation{
			Code:    "dev_skip_validation",
			Message: "CMDCTRL_DEV_SKIP_DECK_VALIDATION is set; deck validation was bypassed",
		})
	} else if verr := deck.Validate(list); verr != nil {
		var ve *deck.ValidationError
		if errors.As(verr, &ve) {
			for _, v := range ve.Violations {
				if v.Code == deck.CodeSideboardUnsupported {
					warnings = append(warnings, v)
					continue
				}
				fatal = append(fatal, v)
			}
			if len(fatal) > 0 {
				return writeDeckViolations(w, "deck has validation errors", fatal, warnings)
			}
		} else {
			return verr
		}
	}

	meta, err := c.Lobby.SetDeck(id, body.PlayerID, list.Name, list.ToGameCards())
	if err != nil {
		return err
	}

	commanders := make([]string, 0, len(list.Commanders))
	for _, cc := range list.Commanders {
		commanders = append(commanders, cc.Name)
	}
	return writeJSON(w, http.StatusOK, uploadDeckResponse{
		Game:       meta,
		DeckName:   list.Name,
		CardCount:  len(list.Commanders) + len(list.Mainboard),
		Commanders: commanders,
		Warnings:   warnings,
	})
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

// writeDeckViolations is the canonical 422 body for deck-upload
// failures: a human-readable summary, the list of fatal violations,
// and any non-fatal warnings (e.g. sideboard contents). All three
// failure modes — unknown card, unsupported mechanic, validation —
// share this schema so the client has one shape to render.
func writeDeckViolations(w http.ResponseWriter, summary string, fatal, warnings []deck.Violation) error {
	body := map[string]any{
		"error":      summary,
		"violations": fatal,
	}
	if len(warnings) > 0 {
		body["warnings"] = warnings
	}
	return writeJSON(w, http.StatusUnprocessableEntity, body)
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
		errors.Is(err, ErrDeckNotUploaded),
		errors.Is(err, game.ErrNotEnoughPlayers),
		errors.Is(err, game.ErrGameAlreadyStarted),
		errors.Is(err, game.ErrGameNotInLobby):
		status = http.StatusConflict
	case errors.Is(err, ErrPlayerNotInGame):
		status = http.StatusForbidden
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
