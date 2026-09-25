package lobby

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/bugstore"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckrequests"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/appenv"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/envflag"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/ratelimit"
)

// sessionTTL controls how long a newly-minted Principal lives in the
// authenticator. Short enough that leaked cookies rotate out of risk
// within a day; long enough that a friends' game session doesn't hit
// "please log in again" mid-match.
const sessionTTL = 12 * time.Hour

// identityTTL is the default lifetime of a RoleIdentified session, the
// one the Discord callback mints when no invite is in hand (ADR 0051
// decision 3: "long TTL (30 days, CMDCTRL_IDENTITY_TTL)"). It is long
// because it is the credential a browser keeps across games; it is
// safe to be long because it carries a UserID and so can be revoked
// (decision 6). Every other session keeps SessionTTL.
const identityTTL = 30 * 24 * time.Hour

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
	// Env is the deployment identity (prod / dev). The zero value is
	// the empty string, which IsDev() reports false for — so a Config
	// built without thinking about it (every existing test) gets
	// production behaviour and no dev surfaces.
	Env appenv.Env
	// Features are the dev-only capabilities this deployment exposes.
	// Always the zero value in production; see package appenv.
	Features   appenv.Features
	SessionTTL time.Duration
	// IdentityTTL is the lifetime of the RoleIdentified session minted
	// at Discord sign-in (CMDCTRL_IDENTITY_TTL). Zero means identityTTL,
	// 30 days. Seat, spectator and admin sessions use SessionTTL.
	IdentityTTL time.Duration
	AllowAnon   bool // allow unauthenticated /games/{id}/join via invite (default: true)
	// Cards is the Scryfall index used by the deck-upload endpoint.
	// When nil, POST /games/{id}/decks returns 503 so a fresh
	// deployment (no Scryfall dump yet) surfaces a clear "run
	// scryfall-refresh.sh" error rather than a cryptic unknown-card
	// list.
	Cards *cards.Index
	// Tokens is the catalog's token templates, for the table spawner
	// (ADR 0075 §2.4). effects.Tokens() satisfies it; main wires it.
	// Nil is supported: POST /games/{id}/spawn still spawns real
	// cards and the token list comes back empty, which is what a
	// deployment or a test that never wired the catalog should see.
	Tokens TokenTemplates
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
	// http.DefaultClient. The DM-invite route shares it; nil there
	// means a client with a short timeout.
	DiscordHTTPClient *http.Client
	// DiscordBot is the server's bot-token credential, for ADR 0051
	// decision 5's direct-message invites (S34 sub-PR 6). The zero
	// value means CMDCTRL_DISCORD_BOT_TOKEN is unset, which is a
	// supported state: POST /games/{id}/invites/dm answers 503 naming
	// the variable and EVERY OTHER ROUTE IS UNCHANGED. It never falls
	// open — an unset token cannot send a DM by another path — and
	// main.go says so once at boot.
	DiscordBot discord.Bot
	// InviteBaseURL is the origin an invite link is built against
	// (CMDCTRL_PUBLIC_BASE_URL, falling back to
	// CMDCTRL_CLIENT_BASE_URL — the same pair bug-report attachments
	// use). "" means the DM route has no link to send and answers
	// 503, for the same reason the bug store has no default: a wrong
	// origin produces a DM full of dead links, which is worse than a
	// deployment where the button simply is not offered.
	InviteBaseURL string
	// Users records who signed in (ADR 0051 decision 2, S34 sub-PR 2).
	// The OAuth callback upserts the Discord identity here and stamps
	// the returned user id onto the session as Principal.UserID. Nil
	// behaves as users.NoStore: sign-in works and the session carries
	// a zero UserID, which is what a deployment with no database gets.
	Users users.Store
	// Revocations is ADR 0051 decision 6's per-user revocation: the
	// write side behind POST /logout/everywhere and POST
	// /admin/users/{id}/revoke-sessions. The read side is not here; it
	// is Auth, which main wraps with auth.WithRevocation over the same
	// *users.Revocations. Nil (no database): the admin route answers
	// 503. Logout-everywhere answers 403 first, because with no
	// database no session has a user, and 503 only if one somehow does.
	Revocations SessionRevoker
	// SessionEvictor closes the WebSockets a revoked user still has
	// open. Nil leaves them up until they next reconnect, when the
	// upgrade is refused.
	SessionEvictor UserSessionEvictor

	// DeckLibrary is a signed-in player's saved decks (ADR 0051
	// decision 7, S34 sub-PR 5). POST /games/{id}/decks saves or
	// updates a row here for a caller with a non-zero
	// Principal.UserID; POST /games/{id}/decks/{deck_id} and
	// GET /me/decks read it back. Nil behaves as
	// decklibrary.NoStore: since Users nil (or unconfigured) already
	// means every principal carries a zero UserID, nothing calls these
	// methods in that deployment shape — this exists so the fallback
	// is total rather than a nil-check the handlers would otherwise
	// need to duplicate.
	DeckLibrary decklibrary.Store

	// DiscordAvatars is the S12.5 avatar cache. Nil means
	// /avatars/* returns 503; production wires a cache rooted at
	// $CMDCTRL_DATA_DIR/avatars so disk-cached images persist
	// across restarts. The client falls back to an initials
	// placeholder when the endpoint 503s, so an unconfigured
	// deploy still renders a usable board.
	DiscordAvatars *discord.AvatarCache

	// BugReporter files in-app bug reports as GitHub issues
	// (bugreport.go). Nil disables the surface: POST /bugreport
	// returns 503 and GET /bugreport/config reports enabled:false
	// so the client hides the report button. Production wires
	// *github.Client when CMDCTRL_GITHUB_TOKEN is set.
	BugReporter BugReporter

	// BugStore persists the artifacts a report carries: reporter
	// screenshots (served publicly so GitHub's image proxy can render
	// them) and a pinned copy of the game's replay (admin-only). Nil
	// or unconfigured means text-only reports: POST /bugreport still
	// files issues, the attachment routes 503, and
	// /bugreport/config reports attachments:false so the modal hides
	// its file picker. See ADR 0017 §6.
	BugStore *bugstore.Store

	// DeckRequestFiler files and joins ADR 0095's deck-request issues.
	// Nil turns POST /deck-requests off (503 naming
	// CMDCTRL_GITHUB_TOKEN). Production wires the same *github.Client
	// as BugReporter.
	DeckRequestFiler DeckRequestFiler

	// DeckRequests remembers which issue tracks each deck and who
	// asked when (migration 0006). Nil turns POST /deck-requests off:
	// the rate limit and the deduplication both live in it.
	DeckRequests deckrequests.Store

	// FetchDeck fetches a deck link for POST /deck-coverage and POST
	// /deck-requests. Nil means deck.FetchFromURL over DeckHTTPClient;
	// tests inject a fake so nothing reaches the network.
	FetchDeck DeckFetcher

	// Log is used for the handful of non-fatal conditions where
	// swallowing the error silently would cost a later debugging
	// session — a bug-report manifest that failed to write, a replay
	// that couldn't be pinned. Nil disables those warnings; nothing
	// in the request path depends on it.
	Log *slog.Logger

	// Bots is the bot runner host (S31). Nil disables the
	// /games/{id}/seats/bot routes with a 503 — the lobby can seat a
	// bot but nothing would ever play it.
	Bots BotHost

	// BotDecks is the named-deck catalog the bot picker offers — the
	// four curated archetypes (cmd/server's botDeckCatalog) in
	// production.
	// Nil means the picker lists no decks and `deck` is refused; the
	// raw-decklist path still works.
	BotDecks aiseat.DeckSource
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
//	GET  /games/{id}/preview?t=<invite> — table name / state / seats for an invite holder
//	POST /games/{id}/start  — authenticated: transition lobby → active
//	POST /games/{id}/archive — admin: retire the table from the listing
//	DELETE /games/{id}/archive — admin: put it back
//	POST /games/{id}/seats/{player}/reclaim — admin: mint a seat-reclaim link
//	POST /games/{id}/reclaim — redeem one: a session for that seat
//	POST /games/{id}/invites/rotate — admin: revoke + re-mint one invite kind
//	POST /games/{id}/invites/dm — seated / creator / admin: DM a person the game's invite link
//	GET  /decks             — authenticated: pre-built decks + their engine coverage
//	POST /games/{id}/decks/{deck_id} — authenticated: seat a library deck without re-pasting
//	GET  /me                — authenticated: principal echo (for client bootstrap)
//	GET  /me/decks          — authenticated: the caller's deck library
//	GET  /me/tablemates     — signed in: the people you have shared a table with
//	POST /logout            — revoke the caller's session server-side
//	POST /logout/everywhere — withdraw every session the caller's user holds
//	POST /admin/users/{id}/revoke-sessions — admin: the same, for any user
//
// Routes that mutate state accept JSON bodies; read-only routes use
// query params / path params. All responses are JSON.
func Handler(c Config) http.Handler {
	if c.SessionTTL == 0 {
		c.SessionTTL = sessionTTL
	}
	if c.IdentityTTL == 0 {
		c.IdentityTTL = identityTTL
	}
	mux := http.NewServeMux()

	// Dev / e2e knob: with CMDCTRL_DEV_RELAX_RATE_LIMITS set (truthy),
	// every limiter constructed here becomes effectively unlimited so
	// test suites can hammer join/login without outwaiting the refill.
	// Same family as CMDCTRL_DEV_SKIP_DECK_VALIDATION; main.go warns
	// at startup when it's active. NEVER set this in production.
	relaxed := envflag.Truthy(os.Getenv("CMDCTRL_DEV_RELAX_RATE_LIMITS"))
	newLimiter := func(rate, burst float64) *ratelimit.Limiter {
		if relaxed {
			return ratelimit.New(1<<20, 1<<20)
		}
		return ratelimit.New(rate, burst)
	}

	// Rate-limit the two endpoints that accept untrusted credentials:
	// /admin/login brute-forces the shared admin token, /join
	// brute-forces invite tokens. 1 req/s with a 5-token burst per
	// client IP is lenient for real humans, prohibitive for scripts.
	limit := newLimiter(1, 5)
	// Deck upload has a separate, more permissive bucket: legitimate
	// players may re-upload several times while iterating, but we still
	// want a ceiling on how fast a single IP can stream megabyte-sized
	// Moxfield blobs at the parser.
	deckLimit := newLimiter(2, 10)
	// Avatar fetches: a board render needs at most one request per
	// seat (the response is cached immutable for a day), so a small
	// steady rate caps how fast one session can make us hit Discord's
	// CDN and write to disk. The burst is sized for several boards
	// cold-loading in the same second THROUGH ONE BUCKET: without
	// CMDCTRL_TRUST_FORWARDED, every client behind a reverse proxy
	// keys to the proxy's address, and a 4-player game is up to 4
	// avatars per viewer.
	avatarLimit := newLimiter(5, 40)
	mux.Handle("POST /admin/login", limit.Middleware(handlerFunc(c, adminLogin)))
	mux.Handle("POST /games/{id}/join", limit.Middleware(handlerFunc(c, joinGame)))
	// Bare-code join (the login-page flow): no {id} in the path
	// because the invite token names the table by itself, via
	// Lobby.FindByInvite. Same limiter bucket as the path-scoped
	// join — it is the same credential being guessed, and a
	// code-only endpoint is the easier of the two to script.
	mux.Handle("POST /join", limit.Middleware(handlerFunc(c, joinByCode)))
	mux.Handle("POST /games/{id}/spectate", limit.Middleware(handlerFunc(c, spectateGame)))
	// Invite-page preview: same bucket as join, since the invite in
	// the query string is the credential being brute-forced.
	mux.Handle("GET /games/{id}/preview", limit.Middleware(handlerFunc(c, previewGame)))

	// Discord OAuth (S12.5). All three are unauthenticated — they
	// either run before any session exists or carry their own
	// CSRF protection via the state token. /start and /callback
	// are rate-limited alongside admin-login + join because state
	// generation involves crypto/rand and the upstream calls are
	// the most expensive thing we forward to Discord.
	// GET /config — deployment identity + dev feature flags. Sits
	// alongside the other unauthenticated capability probes
	// (/auth/discord/config, /bugreport/config) and is fetched by the
	// client before login so the env banner renders on the login
	// screen. Constant and secret-free in production.
	mux.Handle("GET /config", handlerFunc(c, clientConfig))
	mux.Handle("GET /auth/discord/config", handlerFunc(c, discordConfig))
	mux.Handle("GET /auth/discord/start", limit.Middleware(handlerFunc(c, discordStart)))
	mux.Handle("GET /auth/discord/callback", limit.Middleware(handlerFunc(c, discordCallback)))
	// Linking Discord to a seat already held (S34 sub-PR 4, from S12.5
	// #59). Unlike /start it needs a player session: the seat it links
	// is the session's own. Same bucket as /start.
	mux.Handle("GET /auth/discord/link", limit.Middleware(auth.Middleware(c.Auth, auth.RolePlayer)(handlerFunc(c, discordLink))))
	// Discord avatar cache. Session-gated: the board's <img> tags are
	// same-origin, so the httpOnly session cookie rides along without
	// the client attaching a token. Rate-limited because each cold
	// miss costs an outbound CDN fetch plus a disk write — without a
	// ceiling the endpoint is an unmetered write-to-disk-forever
	// primitive for anyone holding a session.
	mux.Handle("GET /avatars/{id}/{hash}", avatarLimit.Middleware(auth.Middleware(c.Auth)(handlerFunc(c, discordAvatar))))
	mux.Handle("POST /games", auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, createGame)))
	mux.Handle("DELETE /games/{id}", auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, deleteGame)))
	// Archive / unarchive: the reversible half of DELETE. Admin-only
	// on the same gate, because hiding somebody else's table from the
	// listing is an operator action even though it destroys nothing.
	mux.Handle("POST /games/{id}/archive", auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, archiveGame)))
	mux.Handle("DELETE /games/{id}/archive", auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, unarchiveGame)))
	// Seat reclaim (see reclaim.go). Two halves with deliberately
	// different gates: minting is ADMIN-ONLY — the ticket is a bearer
	// credential for one player's seat, hidden information and all —
	// while redemption is unauthenticated by necessity (the person
	// redeeming has lost their session; the ticket IS the credential)
	// and therefore rides the same brute-force bucket as join and
	// spectate.
	mux.Handle("POST /games/{id}/seats/{player}/reclaim",
		auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, mintSeatReclaim)))
	mux.Handle("POST /games/{id}/reclaim", limit.Middleware(handlerFunc(c, redeemSeatReclaim)))
	// Invite rotation (#1038, ADR 0051 decision 4): revoke a game's
	// current invite of one kind and mint its replacement, so a link
	// lost to a restart (only its hash survives one — see
	// resolveInvite) can be replaced without deleting and recreating
	// the table. Host or admin (#1098) — "host" here means the game's
	// CREATOR (games.created_by), not the seated table host of
	// host.go/ADR 0075; authorised inside the handler with
	// CanRotateInvites, same pattern as /host's CanManageTable below.
	// Session-gated here (any authenticated role) rather than
	// RoleAdmin so a signed-in creator who hasn't claimed a seat can
	// reach the handler at all.
	// Rate-limited in the join/spectate/preview bucket: it mints and
	// revokes the exact credential those routes brute-force, even
	// though the auth gate already keeps a stranger out.
	mux.Handle("POST /games/{id}/invites/rotate",
		limit.Middleware(auth.Middleware(c.Auth)(handlerFunc(c, rotateInvite))))
	// Direct-message invites (ADR 0051 decision 5, S34 sub-PR 6): the
	// server DMs a tablemate the game's ORDINARY player invite through
	// Discord's REST API. Session-gated, then authorised in the
	// handler (seated at the table, the game's creator, or admin).
	//
	// Two buckets, both of which have to allow the request. The IP
	// bucket is the one every other invite-adjacent route rides. The
	// per-CALLER bucket is what decision 5 actually asks for: every
	// accepted call costs two outbound Discord writes and lands an
	// unsolicited DM in somebody's inbox, and an IP bucket is shared
	// by everyone behind one reverse proxy — it would let one tab
	// spend the whole table's allowance. ~1 DM / 10 s with a burst of
	// 3 is generous for inviting three friends at once and useless
	// for anything else.
	dmLimit := newLimiter(1.0/10, 3)
	mux.Handle("POST /games/{id}/invites/dm",
		limit.Middleware(auth.Middleware(c.Auth)(perCallerLimit(dmLimit, handlerFunc(c, inviteDM)))))
	// The Discord bot's /c2-end host check (#1098): "does this Discord
	// user's snowflake match this game's creator", a boolean and
	// nothing else. Admin-only — the bot always calls with its admin
	// session — so it never widens who can learn a game's creator;
	// GET /games/{id} never serves the raw creator identity to
	// anyone, admin included (see redactMetaFor), and this route
	// exists so the bot can ask the one question it actually needs
	// without that identity ever leaving the server.
	mux.Handle("GET /games/{id}/creator",
		auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, gameCreator)))
	mux.Handle("GET /games", auth.Middleware(c.Auth)(handlerFunc(c, listGames)))
	mux.Handle("GET /games/{id}", auth.Middleware(c.Auth)(handlerFunc(c, getGame)))
	mux.Handle("POST /games/{id}/start", auth.Middleware(c.Auth)(handlerFunc(c, startGame)))
	// ADR 0075 §2.1: hand the table to another seat. Host or admin;
	// authorised inside the handler with CanManageTable.
	mux.Handle("POST /games/{id}/host", auth.Middleware(c.Auth)(handlerFunc(c, transferHost)))
	// ADR 0075 §2.3: change the table's settings. Host or admin,
	// authorised inside the handler with CanManageTable. PATCH
	// because the body is a partial — an absent field is "leave it
	// alone", which is the difference between turning spawning on and
	// resetting the undo limit as a side effect. Accepted in the
	// lobby and on a live game alike; the per-field timing rules
	// (starting life is fixed once the game is active) are the
	// engine's, not the route's.
	mux.Handle("PATCH /games/{id}/settings", auth.Middleware(c.Auth)(handlerFunc(c, updateTableSettings)))
	// ADR 0075 §2.4: spawning on a LIVE table. Deliberately NOT
	// behind requireDevFeature — this is the production spawner, and
	// its two gates are its own: CanManageTable, and the table's
	// AllowSpawn setting (off by default). Both are checked in the
	// handler, and a refusal says which one fired. The dev route
	// below keeps its dev-only, anyone-at-the-table semantics.
	mux.Handle("POST /games/{id}/spawn", auth.Middleware(c.Auth)(handlerFunc(c, spawnCard)))
	mux.Handle("GET /games/{id}/spawn/tokens", auth.Middleware(c.Auth)(handlerFunc(c, spawnTokens)))
	// The card picker's search. GET /dev/cards is the same read, but
	// it 404s in production, so the production spawner needs its own
	// — scoped to a game so it rides the /games prefix every proxy
	// already carries, and gated like the spawn it feeds.
	mux.Handle("GET /games/{id}/spawn/cards", auth.Middleware(c.Auth)(handlerFunc(c, spawnCardSearch)))
	mux.Handle("GET /games/{id}/replay", auth.Middleware(c.Auth)(handlerFunc(c, downloadReplay)))
	// S15 sub-PR 4 — read-only auto-tap preview. The client polls
	// this just before firing cast_spell with auto_tap=true; the
	// response shape is the planned tap order so the cast modal
	// can show the user which permanents will tap before they
	// confirm. Read-only — no game state mutates.
	mux.Handle("GET /games/{id}/auto-tap-preview", auth.Middleware(c.Auth)(handlerFunc(c, autoTapPreview)))
	mux.Handle("POST /games/{id}/decks", deckLimit.Middleware(auth.Middleware(c.Auth)(handlerFunc(c, uploadDeck))))
	// Seat a deck already in the caller's library (ADR 0051 decision
	// 7, S34 sub-PR 5) without re-pasting it. Same rate bucket as
	// /decks — it re-parses and re-validates a decklist just like an
	// upload does.
	mux.Handle("POST /games/{id}/decks/{deck_id}", deckLimit.Middleware(auth.Middleware(c.Auth)(handlerFunc(c, seatLibraryDeck))))
	// S31: bot seats. Same deck pipeline (and the same rate bucket —
	// the body is a decklist) as /decks; authorised for admin or any
	// player already seated at the table.
	mux.Handle("POST /games/{id}/seats/bot", deckLimit.Middleware(auth.Middleware(c.Auth)(handlerFunc(c, addBot))))
	mux.Handle("DELETE /games/{id}/seats/bot/{player}", auth.Middleware(c.Auth)(handlerFunc(c, removeBot)))
	// #505 part 3: admin-only latency/token readout for every bot seat
	// at this table. BotStatsHandler wraps its own
	// auth.Middleware(c.Auth, auth.RoleAdmin) — see botstats.go — so
	// this is the one line that mounts it, matching every other admin
	// route above.
	mux.Handle(BotStatsRoute, BotStatsHandler(c))
	// What the picker needs before it can offer anything: the tier
	// list (including the ones that are declared but not built, so
	// the UI can grey them out) and the curated deck catalog.
	// Session-gated but game-independent — it is the same answer for
	// every table. NOTE: /bot is a new top-level prefix; it is in
	// deploy/Caddyfile's @api matcher, client/vite.config.ts and the
	// service worker's API_PATH, all of which have to agree or this
	// 404s only in production.
	mux.Handle("GET /bot/options", auth.Middleware(c.Auth)(handlerFunc(c, botOptions)))
	// The human deck picker's catalog: the same pre-built decks the
	// bot picker offers, plus each one's engine-coverage profile.
	// Session-gated and game-independent, like /bot/options, and
	// deliberately NOT folded into it — a deployment with no bot host
	// still wants a deck picker. Same top-level-prefix warning as
	// /bot: /decks has to be in deploy/Caddyfile's @api matcher,
	// client/vite.config.ts's proxy and the service worker's
	// API_PATH, or it 404s in production only.
	mux.Handle("GET /decks", auth.Middleware(c.Auth)(handlerFunc(c, prebuiltDecks)))
	mux.Handle("GET /me", auth.Middleware(c.Auth)(handlerFunc(c, me)))
	// "My games" and seat reclaim by user (ADR 0051 decisions 3 and 4,
	// S34 sub-PR 4). Session-gated here; the handlers then require a
	// UserID on it, since the answer is a person's, not a seat's.
	// /me/* is in deploy/Caddyfile's @api matcher; /me in the Vite
	// proxy and the service worker's API_PATH already covers it.
	mux.Handle("GET /me/games", auth.Middleware(c.Auth)(handlerFunc(c, myGames)))
	// The invite picker's list (ADR 0051 decision 8, S34 sub-PR 6):
	// the people the caller has shared a table with. Same caller rule
	// as the rest of /me/*.
	mux.Handle("GET /me/tablemates", auth.Middleware(c.Auth)(handlerFunc(c, myTablemates)))
	mux.Handle("POST /me/games/{id}/session", auth.Middleware(c.Auth)(handlerFunc(c, myGameSession)))
	// The caller's deck library (ADR 0051 decision 7, S34 sub-PR 5).
	// Any authenticated role reaches the handler; it 401s itself for a
	// principal with no UserID (a guest, an admin, or an identified
	// session from a no-database deployment).
	mux.Handle("GET /me/decks", auth.Middleware(c.Auth)(handlerFunc(c, myDecks)))

	// Develop-environment card spawner (ADR 0023). Both routes are
	// wrapped in requireDevFeature: in production they are 404s, and
	// they stay 404s on a dev deployment that has switched the
	// card_spawn feature off. Session-gated but not admin-gated --
	// on a preview box anyone at the table is a tester.
	//
	// The gate is the outermost wrapper on purpose: an unauthenticated
	// probe against production must not be able to tell these apart
	// from any other unrouted path, and auth.Middleware would answer
	// 401 first and confirm the route exists.
	devSpawn := func(h http.Handler) http.Handler {
		return requireDevFeature(c, func(f appenv.Features) bool { return f.CardSpawn }, h)
	}
	mux.Handle("GET /dev/cards", devSpawn(auth.Middleware(c.Auth)(handlerFunc(c, devCardSearch))))
	mux.Handle("POST /games/{id}/dev/spawn", devSpawn(auth.Middleware(c.Auth)(handlerFunc(c, devSpawnCard))))
	// Bug reports (in-app button → GitHub issue). The config probe
	// is unauthenticated and mirrors /auth/discord/config. POST is
	// session-gated (any role — spectators hit bugs too) with its
	// own tight bucket: every accepted report costs an outbound
	// GitHub write, and one stuck retry loop shouldn't be able to
	// wallpaper the tracker. ~1 report / 30 s with a burst of 3.
	bugLimit := newLimiter(1.0/30, 3)
	// Attachment reads get their own, much looser bucket. This route
	// is fetched by GitHub's Camo proxy (once per image, from GitHub's
	// address space) and then by whoever opens the issue, so the tight
	// filing limit would starve legitimate renders. The limit here is
	// only a brake on enumeration attempts, which 404 anyway.
	bugAttachLimit := newLimiter(5, 60)
	// ADR 0095: the deck coverage checker and deck requests.
	//
	// POST /deck-coverage is public and fetches a third-party URL on
	// the caller's behalf, so it is limited per client IP: ~1 check /
	// 10 s with a burst of 3. The bot's admin session rides its own
	// bucket instead, because it calls from loopback for every guild
	// member at once (deckCoverageLimit). A repeat check of one deck
	// is served from a ten-minute cache and fetches nothing.
	//
	// POST /deck-requests needs a session and rides the ordinary IP
	// bucket; its real limit is three asks per requester per 24 h,
	// counted in the database.
	deckChecks := newDeckCheck()
	deckCoveragePublic := newLimiter(1.0/10, 3)
	deckCoverageAdmin := newLimiter(1, 10)
	mux.Handle("POST /deck-coverage",
		deckCoverageLimit(c, deckCoveragePublic, deckCoverageAdmin, handlerFunc(c, deckChecks.coverage)))
	mux.Handle("POST /deck-requests",
		limit.Middleware(auth.Middleware(c.Auth)(handlerFunc(c, deckChecks.request))))

	mux.Handle("GET /bugreport/config", handlerFunc(c, bugReportConfig))
	mux.Handle("POST /bugreport", bugLimit.Middleware(auth.Middleware(c.Auth)(handlerFunc(c, bugReport))))
	// Unauthenticated by necessity — Camo presents no session. The
	// report ID in the path is the capability, and it only ever
	// appears inside a private-repo issue. bugstore enforces the
	// magic-byte allowlist and the inert response headers.
	mux.Handle("GET /bugreport/att/{id}/{name}", bugAttachLimit.Middleware(handlerFunc(c, bugAttachment)))
	// The pinned replay is the raw unfiltered view — admin only, with
	// no game-has-ended relaxation (see bugPinnedReplay).
	mux.Handle("GET /bugreport/{id}/replay", auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, bugPinnedReplay)))
	// The pinned public game log rides the same admin gate as the
	// replay — the contents are public, the artifact is unfiltered.
	mux.Handle("GET /bugreport/{id}/gamelog", auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, bugPinnedGameLog)))
	// Logout does not require an authenticated principal — a client
	// with a stale or revoked token should still be able to clear
	// browser state without a 401 dead-end. We just revoke whatever
	// credential is on the request (if any) and drop the cookie.
	mux.Handle("POST /logout", handlerFunc(c, logout))
	// Per-user revocation (ADR 0051 decision 6, S34 sub-PR 7). Unlike
	// /logout these need a valid session: logout-everywhere acts on
	// the caller's own user, and only an admin may revoke someone
	// else's. /logout/* is its own entry in deploy/Caddyfile's @api
	// matcher, because Caddy's /logout matches that path exactly.
	mux.Handle("POST /logout/everywhere", auth.Middleware(c.Auth)(handlerFunc(c, logoutEverywhere)))
	mux.Handle("POST /admin/users/{id}/revoke-sessions", auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, adminRevokeUserSessions)))

	return mux
}

// handlerFunc adapts a (Config, w, r) → error closure into an
// http.Handler, centralising error-to-JSON translation so every
// endpoint doesn't repeat the same switch statement.
type lobbyHandler func(c Config, w http.ResponseWriter, r *http.Request) error

// perCallerLimit throttles by WHO is calling rather than by where
// from: the caller's user, or their session's seat when they have no
// user, or their admin role. It must sit INSIDE auth.Middleware,
// which is what puts the principal in the request context.
//
// ratelimit.Limiter.Middleware keys on the client IP, which is the
// right key for a credential being brute-forced and the wrong one for
// a route whose cost is per person: without CMDCTRL_TRUST_FORWARDED
// every caller behind a reverse proxy shares one bucket, so an IP
// bucket would let one tab spend the whole table's allowance. The two
// compose — a request has to satisfy both.
func perCallerLimit(l *ratelimit.Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(callerKey(r)) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "10")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"too many requests"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// callerKey identifies the principal for perCallerLimit. Every branch
// is prefixed so a user id can never collide with a player id.
func callerKey(r *http.Request) string {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return "anon"
	}
	switch {
	case p.UserID != uuid.Nil:
		return "user:" + p.UserID.String()
	case p.Role == auth.RoleAdmin:
		// One bucket for the admin credential, shared by the bot and
		// any operator holding it. That is the point: it is one
		// credential.
		return "admin"
	case p.PlayerID != uuid.Nil:
		return "seat:" + p.PlayerID.String()
	default:
		return "anon"
	}
}

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
	// HostDiscordID optionally names the table host by Discord user
	// ID (ADR 0075 §2.1). /c2-invite sends the invoking user.
	HostDiscordID string `json:"host_discord_id,omitempty"`
}

// transferHostRequest is the body of POST /games/{id}/host.
type transferHostRequest struct {
	PlayerID uuid.UUID `json:"player_id"`
}

// tableSettingsResponse is the body of PATCH /games/{id}/settings:
// the table's settings AFTER the patch, whole. Wrapped in an object
// rather than returned bare so the route has somewhere to grow (the
// ADR's §2.3 timing notes are a natural second field) without
// breaking a client that already reads `settings`.
type tableSettingsResponse struct {
	Settings protocol.TableSettingsView `json:"settings"`
}

// reclaimRequest is the body of POST /games/{id}/reclaim. Ticket is
// the whole credential; there is deliberately no name field.
type reclaimRequest struct {
	Ticket string `json:"ticket"`
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
	if err := decodeJSON(w, r, &body); err != nil {
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
	setSessionCookie(c, w, tok, issued.ExpiresAt)
	return writeJSON(w, http.StatusOK, sessionResponse{Token: tok, ExpiresAt: issued.ExpiresAt, Principal: issued})
}

// createGame (admin-only) creates a new game in the lobby and
// returns its metadata INCLUDING the invite token. The admin is
// responsible for distributing the invite out-of-band.
//
// games.created_by is the caller's UserID (ADR 0051 decision 2). The
// route is admin-only, and an admin session carries no UserID, so
// today that is always NULL; it is read from the principal rather
// than hard-coded so opening the route to signed-in users later is a
// change to the middleware line and nothing here.
func createGame(c Config, w http.ResponseWriter, r *http.Request) error {
	var body createGameRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	p, _ := auth.PrincipalFromContext(r.Context())
	meta, err := c.Lobby.CreateWith(body.Name, p.UserID, body.HostDiscordID)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusCreated, redactMetaFor(p, meta.ID, meta))
}

// transferHost handles POST /games/{id}/host: hand the table to
// another seat (ADR 0075 §2.1). Host or admin only; the target must
// be a human seat still in this game.
func transferHost(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	var body transferHostRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	if !CanManageTable(p, meta) {
		return ErrNotTableManager
	}
	if body.PlayerID == uuid.Nil {
		return httpError(http.StatusBadRequest, "player_id is required")
	}
	meta, err = c.Lobby.TransferHost(id, body.PlayerID)
	if errors.Is(err, ErrPlayerNotInGame) {
		// The CALLER is fine; the target is not a seat here.
		return httpError(http.StatusUnprocessableEntity, "player_id is not a seat in this game")
	}
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, redactMetaFor(p, id, meta))
}

// updateTableSettings handles PATCH /games/{id}/settings: change the
// table's rules (ADR 0075 §2.3). Host or admin only.
//
// The body IS a game.SettingsPatch — {"undo_limit": 3} changes the
// undo limit and nothing else. Decoding straight into the engine's
// patch type is deliberate: a lobby-side mirror of six pointer fields
// would be six chances for the wire name and the engine's to drift,
// and the wire names are already pinned by the JSON tags the game view
// publishes (protocol.TableSettingsView).
//
// The caller's own seat rides onto the event as the actor, so the log
// names them. An admin session has no seat and passes uuid.Nil, which
// the log renders as "The admin".
func updateTableSettings(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	var patch game.SettingsPatch
	if err := decodeJSON(w, r, &patch); err != nil {
		return err
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	if !CanManageTable(p, meta) {
		return ErrNotTableManager
	}
	// An admin session carries no PlayerID, which is exactly the
	// uuid.Nil the engine records for "the server admin".
	settings, err := c.Lobby.UpdateSettings(id, p.PlayerID, patch)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, tableSettingsResponse{
		Settings: protocol.ViewOfTableSettings(settings),
	})
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
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}

	// A signed-in person clicking an invite link sits as themselves
	// (ADR 0051 sub-PR 4): the seat takes the Discord identity and the
	// user from the session, and body.name is ignored, as on POST
	// /join. Anyone else, including a session that no longer
	// validates, joins exactly as before, by name.
	identity, userID := signedInIdentity(c, r)
	name := body.Name
	if identity.Populated() {
		name = ""
	}
	meta, playerID, err := c.Lobby.JoinAs(id, body.InviteToken, name, identity, userID)
	if err != nil {
		return err
	}

	// Mint a RolePlayer session bound to (gameID, playerID). The WS
	// authorizer will cross-check the principal's GameID against the
	// one on the upgrade request — a player session can't be reused
	// to spy on a different game.
	p := auth.Principal{
		Role:     auth.RolePlayer,
		UserID:   userID,
		GameID:   meta.ID,
		PlayerID: playerID,
		Name:     body.Name,
	}
	if identity.Populated() {
		p.Name = identity.DisplayName()
		p.DiscordID = identity.ID
		p.DiscordUsername = identity.Username
		p.DiscordGlobalName = identity.GlobalName
		p.DiscordAvatarHash = identity.AvatarHash
	}
	tok, issued, err := c.Auth.Issue(r.Context(), p, c.SessionTTL)
	if err != nil {
		return err
	}
	setSessionCookie(c, w, tok, issued.ExpiresAt)

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

// joinByCode is the login-page counterpart to joinGame: the caller
// holds an invite code but no game id, which is the shape you get
// when somebody pastes a code out of Discord chat instead of
// clicking a link.
//
// Three callers, told apart by the session attached (if any):
//
//   - RoleIdentified — a Discord sign-in from the login page. The
//     seat takes its name and avatar from the session's Discord
//     identity and `name` in the body is ignored, so the identity
//     on the seat is the one Discord vouched for rather than
//     whatever the body claimed.
//   - No session — behaves exactly like the classic join and
//     requires `name`. Manual entry stays first-class: a deploy
//     with Discord unconfigured still has to work.
//   - Any other role — refused. An admin or an already-seated
//     player arriving here is a client bug, and quietly minting
//     them a second seat is worse than an error.
//
// The identity session is deliberately left valid after the swap.
// It is how the same person joins a second table later without
// signing in to Discord again; it cannot do anything on its own
// (the WS authorizer refuses it outright).
func joinByCode(c Config, w http.ResponseWriter, r *http.Request) error {
	var body joinRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	gameID, err := c.Lobby.FindByInvite(body.InviteToken)
	if err != nil {
		return err
	}

	// The session is OPTIONAL on this route, so a credential that
	// fails to validate falls through to the manual-name path
	// rather than 401-ing. An identity session that expired while
	// the user was hunting for the code should feel like "type your
	// name", not like being thrown out.
	var (
		identity DiscordIdentity
		userID   uuid.UUID
	)
	if cred := auth.CredentialFromRequest(r); cred != "" {
		p, verr := c.Auth.Validate(r.Context(), cred)
		switch {
		case verr != nil:
			// Expired or bogus — treat the caller as anonymous.
		case p.Role == auth.RoleIdentified:
			// The seat session is minted from the identity session,
			// so it belongs to the same person (ADR 0051 decision 3).
			userID = p.UserID
			identity = DiscordIdentity{
				ID:         p.DiscordID,
				Username:   p.DiscordUsername,
				GlobalName: p.DiscordGlobalName,
				AvatarHash: p.DiscordAvatarHash,
			}
		default:
			return httpError(http.StatusConflict,
				"this session already belongs to a table — sign out before joining another")
		}
	}

	// An empty name lets JoinWithIdentity fall back to the Discord
	// display name; passing the body's name alongside a verified
	// identity would let a caller relabel a seat Discord vouched for.
	name := body.Name
	if identity.Populated() {
		name = ""
	}
	meta, playerID, err := c.Lobby.JoinAs(gameID, body.InviteToken, name, identity, userID)
	if err != nil {
		return err
	}

	p := auth.Principal{
		Role:     auth.RolePlayer,
		UserID:   userID,
		GameID:   meta.ID,
		PlayerID: playerID,
		Name:     name,
	}
	if identity.Populated() {
		p.Name = identity.DisplayName()
		p.DiscordID = identity.ID
		p.DiscordUsername = identity.Username
		p.DiscordGlobalName = identity.GlobalName
		p.DiscordAvatarHash = identity.AvatarHash
	}
	tok, issued, err := c.Auth.Issue(r.Context(), p, c.SessionTTL)
	if err != nil {
		return err
	}
	setSessionCookie(c, w, tok, issued.ExpiresAt)

	// Same scrub as joinGame: the joiner already has the player
	// invite, and the spectator invite is admin-only data.
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

// previewResponse is the body of GET /games/{id}/preview.
type previewResponse struct {
	Game     GameMeta    `json:"game"`
	Invite   PreviewKind `json:"invite"`
	MaxSeats int         `json:"max_seats"`
}

// previewGame lets an invite holder see the table before joining:
// GET /games/{id}/preview?t=<invite>. Unauthenticated — the invite
// is the credential — and rate-limited with join. The meta comes
// back scrubbed (see Lobby.Preview).
func previewGame(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	meta, kind, err := c.Lobby.Preview(id, r.URL.Query().Get("t"))
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, previewResponse{
		Game:     meta,
		Invite:   kind,
		MaxSeats: game.MaxPlayers,
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
	if err := decodeJSON(w, r, &body); err != nil {
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
	setSessionCookie(c, w, tok, issued.ExpiresAt)

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

// listGames returns the active tables. `?archived=1` returns the
// retired ones instead — a separate view rather than a mixed list,
// so the default answer to "what tables are there" never grows
// without bound as old games pile up.
//
// Each entry goes through redactMetaFor so its is_creator bit (#1098)
// is computed for THIS caller — the lobby UI's rotate-invite buttons
// read it straight off the list, the same response the page already
// polls, rather than a second per-game round trip.
func listGames(c Config, w http.ResponseWriter, r *http.Request) error {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	games := c.Lobby.List()
	if envflag.Truthy(r.URL.Query().Get("archived")) {
		games = c.Lobby.ListArchived()
	}
	for i := range games {
		games[i] = redactMetaFor(p, games[i].ID, games[i])
	}
	return writeJSON(w, http.StatusOK, listResponse{Games: games})
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
	return writeJSON(w, http.StatusOK, redactMetaFor(p, id, meta))
}

// redactMetaFor strips what principal p may not read off a game's
// meta: the invite tokens for anyone outside the table, and both
// tokens for a spectator. It also turns the raw CreatedBy (#1098)
// into the per-viewer IsCreator bit and clears the raw field, so
// nothing downstream of this ever hands out the creator's identity —
// only "yes/no, that's you". CreatedBy is already `json:"-"` and so
// never reaches the wire on its own, but every GameMeta a handler
// serializes should still pass through here rather than rely on that
// alone.
func redactMetaFor(p auth.Principal, id uuid.UUID, meta GameMeta) GameMeta {
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
	meta.IsCreator = p.UserID != uuid.Nil && meta.CreatedBy != uuid.Nil && p.UserID == meta.CreatedBy
	meta.CreatedBy = uuid.Nil
	return meta
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

// archiveGame (admin-only) retires a table: it leaves the default
// listing, its bot runners stop, and anyone still connected is
// evicted — but nothing on disk is removed and unarchiveGame puts it
// back exactly as it was. This is the action to reach for when
// clearing out old games; DELETE is the one that cannot be undone.
//
// Evicting is the honest consequence of hiding a live table. It is
// the same close the deleted-game path sends, so a connected client
// renders an ended state rather than redialling a game that is no
// longer listed.
func archiveGame(c Config, w http.ResponseWriter, r *http.Request) error {
	return setGameArchived(c, w, r, true)
}

// unarchiveGame (admin-only) returns a retired table to the active
// listing, relaunching its bot runners if it was mid-game.
func unarchiveGame(c Config, w http.ResponseWriter, r *http.Request) error {
	return setGameArchived(c, w, r, false)
}

func setGameArchived(c Config, w http.ResponseWriter, r *http.Request, archived bool) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	meta, err := c.Lobby.SetArchived(id, archived)
	if err != nil {
		return err
	}
	if archived && c.Evictor != nil {
		c.Evictor.EvictGame(id)
	}
	// The listing already strips these; an archive response is a
	// listing row, not an ownership grant.
	meta.InviteToken = ""
	meta.SpectatorInvite = ""
	return writeJSON(w, http.StatusOK, meta)
}

// reclaimTicketResponse is the body of POST
// /games/{id}/seats/{player}/reclaim. `ticket` is the secret; every
// other field exists so the host can see what they just minted —
// which seat, for how long, and that it only works once — without
// having to read the docs.
type reclaimTicketResponse struct {
	Ticket     string    `json:"ticket"`
	GameID     uuid.UUID `json:"game_id"`
	PlayerID   uuid.UUID `json:"player_id"`
	Seat       int       `json:"seat"`
	PlayerName string    `json:"player_name"`
	ExpiresAt  time.Time `json:"expires_at"`
	TTLSeconds int       `json:"ttl_seconds"`
	SingleUse  bool      `json:"single_use"`
}

// mintSeatReclaim (admin-only) issues a one-shot, short-lived link
// that puts a disconnected player back in their own seat at a table
// that has already started. See reclaim.go for the security shape;
// the only thing this layer adds is the admin gate.
//
// The ticket is returned in the response body and NOWHERE else: it
// is not logged, not persisted, and not written to the game
// metadata.
func mintSeatReclaim(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	playerID, err := uuid.Parse(r.PathValue("player"))
	if err != nil {
		return httpError(http.StatusBadRequest, "invalid player id")
	}
	ticket, err := c.Lobby.MintReclaim(id, playerID)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusCreated, reclaimTicketResponse{
		Ticket:     ticket.Token,
		GameID:     ticket.GameID,
		PlayerID:   ticket.PlayerID,
		Seat:       ticket.Seat,
		PlayerName: ticket.PlayerName,
		ExpiresAt:  ticket.ExpiresAt,
		TTLSeconds: int(ReclaimTTL / time.Second),
		SingleUse:  true,
	})
}

// redeemSeatReclaim is the public half: the disconnected player
// follows the link, the client posts the ticket here, and a session
// bound to (gameID, playerID) comes back. Unauthenticated because
// the caller by definition has no session — the ticket is the
// credential — and rate-limited alongside join / spectate for the
// same reason.
//
// No name is accepted: the seat already has one, and letting the
// redeemer supply a label would make reclaim a way to rename another
// player. The seat's Discord identity is copied onto the principal so
// the avatar renders exactly as it did before the disconnect.
func redeemSeatReclaim(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	var body reclaimRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}

	meta, seat, err := c.Lobby.RedeemReclaim(id, body.Ticket)
	if err != nil {
		return err
	}

	p := auth.Principal{
		Role:              auth.RolePlayer,
		GameID:            meta.ID,
		PlayerID:          seat.PlayerID,
		Name:              seat.Name,
		DiscordID:         seat.DiscordID,
		DiscordGlobalName: seat.DisplayName,
		DiscordAvatarHash: seat.DiscordAvatarHash,
	}
	tok, issued, err := c.Auth.Issue(r.Context(), p, c.SessionTTL)
	if err != nil {
		return err
	}
	setSessionCookie(c, w, tok, issued.ExpiresAt)

	// A reclaimed seat gets a session, not the table's secrets: the
	// holder proved they own one seat, not that they are the host.
	meta.InviteToken = ""
	meta.SpectatorInvite = ""
	return writeJSON(w, http.StatusOK, sessionResponse{
		Token:     tok,
		ExpiresAt: issued.ExpiresAt,
		Principal: issued,
		Game:      &meta,
		PlayerID:  seat.PlayerID,
	})
}

// rotateInviteRequest is the body of POST /games/{id}/invites/rotate.
type rotateInviteRequest struct {
	Kind string `json:"kind"` // "player" | "spectator"
}

// rotateInviteResponse hands back the freshly-minted plaintext once —
// the same "shown here and nowhere else" rule Create and
// mintSeatReclaim follow. `kind` echoes the request so the client
// doesn't have to remember which button it pressed.
type rotateInviteResponse struct {
	Kind  string `json:"kind"`
	Token string `json:"token"`
}

// rotateInvite (the game's creator or the admin — CanRotateInvites,
// #1098) revokes a game's current invite of one kind and mints its
// replacement in one move: Lobby.RotateInvite. It exists because only
// the process that minted a game can show its invite plaintext (ADR
// 0051 decision 4), so a link lost to a restart could never be
// recovered before this route — the table just sat there with an
// invite nobody could read or reissue. The OLD link of that kind
// stops working the instant this returns; the other kind is
// untouched.
func rotateInvite(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	var body rotateInviteRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	if !CanRotateInvites(p, meta) {
		return ErrNotInviteManager
	}
	newToken, _, err := c.Lobby.RotateInvite(id, InviteKind(body.Kind))
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, rotateInviteResponse{Kind: body.Kind, Token: newToken})
}

// gameCreatorResponse is the body of GET /games/{id}/creator.
type gameCreatorResponse struct {
	IsCreator bool `json:"is_creator"`
}

// gameCreator (admin-only) answers the Discord bot's /c2-end host
// check (#1098): does the Discord user named by ?discord_id=<snowflake>
// match this game's creator. It never says who the creator actually
// is — only whether ONE named snowflake is a match — so an admin
// caller that guesses wrong learns nothing, and nobody who isn't
// already an admin can call this at all.
//
// A game with no creator (games.created_by NULL: admin-created, or
// restored from a pre-ADR-0051 file import) always answers false, for
// every discord_id — there is nothing to match, so the bot's own
// allowlists stay the only route for those games. An unknown
// discord_id (no linked identities row — including every deployment
// with no user database, users.NoStore) answers false the same way
// rather than erroring, for the same "never fail open, never fail
// closed with a 500" reason the two allowlists already follow.
func gameCreator(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	discordID := strings.TrimSpace(r.URL.Query().Get("discord_id"))
	if discordID == "" {
		return httpError(http.StatusBadRequest, "discord_id is required")
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	if meta.CreatedBy == uuid.Nil {
		return writeJSON(w, http.StatusOK, gameCreatorResponse{})
	}
	userID, err := c.userStore().UserIDForDiscord(r.Context(), discordID)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return writeJSON(w, http.StatusOK, gameCreatorResponse{})
		}
		return err
	}
	return writeJSON(w, http.StatusOK, gameCreatorResponse{IsCreator: userID == meta.CreatedBy})
}

// autoTapPreview is the S15 sub-PR 4 read-only auto-tap endpoint.
// The client polls it just before firing cast_spell with
// `auto_tap: true` so the cast modal can show which permanents
// will tap before the user confirms. Read-only — no game state
// mutates. Auth: any seated player (or admin / spectator) at this
// game; the caller's own player ID is read from query params and
// validated against their session principal.
//
// #696: the preview PRICES THE ANNOUNCEMENT, not the card. Every
// query param below that the cast payload carries is passed straight
// into game.CastSpellParams and handed to g.PriceCast, the engine's
// one cast pricer — so a flashback, escape, overload, kicked,
// back-face or granted exile cast is previewed at the price CastSpell
// will charge for it. The endpoint used to parse `card.ManaCost` and
// re-apply the commander tax and the cost modifiers itself, which
// knew nothing about any of those and disabled "Auto-tap & cast" on
// casts that would have gone through.
//
// Query params:
//
//	?card=<instance-uuid>     — required. The card the caller plans
//	                            to cast.
//	?from_zone=<zone>         — optional. The zone the cast comes out
//	                            of: "hand" (the default), "command",
//	                            "graveyard", "exile" or "library".
//	                            Must match the cast_spell payload the
//	                            confirm button will send, because the
//	                            zone decides the commander tax, which
//	                            cost modifiers see the cast, and which
//	                            granted permission prices it.
//	?alternative_cost=<key>   — optional. The CR 118.9 cost the cast
//	                            claims ("flashback", "overload",
//	                            "evoke"). Empty means the printed
//	                            cost. A key the card does not offer
//	                            from that zone is a 400, the same
//	                            refusal the cast gets.
//	?optional_costs=<i>,<i>   — optional. Positions in the card's
//	                            optional additional costs (ADR 0073),
//	                            repeated once per payment for a
//	                            multikicker, so the preview prices the
//	                            kicked cast the caster is announcing.
//	?tap_ids=<uuid>,<uuid>... — optional. Permanents being tapped for
//	                            convoke or waterbend. They pay part of
//	                            the cost, so the plan must not also
//	                            tap them for mana.
//	?sacrifice_ids=<uuid>,...  — optional (#1242). Permanents the
//	?discard_ids=<uuid>,...      cast names to its additional cost's
//	                            sacrifice and discard. They do not
//	                            change the price; the plan must not
//	                            also spend them on mana (a named
//	                            Eldrazi Spawn, a named Spirit Guide).
//	?face=<int>              — optional. The printed face being cast
//	                            (ADR 0034). A modal DFC's back face
//	                            has its own mana cost.
//	?x=<int>                  — optional. Caller-supplied X value
//	                            for spells with {X} in their cost.
//	                            Defaults to 0.
//	?ability=<int>            — optional. Price the card's CR 602
//	                            activated ability at this index
//	                            instead of its printed cast cost,
//	                            so the X picker an {X} ability opens
//	                            reads the ability's own price
//	                            (Helm of Obedience's "{X}", not the
//	                            "{4}" in the card's corner). Every
//	                            cast-shaped param above is ignored on
//	                            this branch — an ability is not a cast
//	                            (tap_ids, sacrifice_ids and discard_ids
//	                            are read, as the activation's own).
//	                            Priced as ActivateCatalogAbility
//	                            charges it (#1405): board modifiers
//	                            and the ability's own clause.
//	?targets=<kind>:<uuid>,…  — optional, with ?ability= only (#1405).
//	                            The announced targets, kind `card` or
//	                            `player`, so a target-keyed price
//	                            (Dragonfire Blade's "{1} less for each
//	                            color of the creature it targets") is
//	                            previewed at the target's price.
//	                            Omitted, the no-target price.
//	?tap_ids= ?sacrifice_ids= — with ?ability= (#1422), optional.
//	?discard_ids= ?exile_ids=    The permanents and cards the
//	?waterbend_ids=              activation names to its TapOthers,
//	                            sacrifice, discard, exile-cards and
//	                            waterbend components, as the
//	                            activate_ability payload names them.
//	                            The plan never spends them, nor the
//	                            source when the cost taps or
//	                            sacrifices it (the set
//	                            ActivateCatalogAbility excludes);
//	                            each waterbend tap also pays {1}.
//	?exclude=<uuid>,<uuid>... — optional. Comma-separated lock-tap
//	                            permanent IDs the auto-tapper must
//	                            NOT consider; lets the client
//	                            reserve sources for later casts.
//
// Response:
//
//	{
//	  "ok": true,
//	  "plan": ["<uuid>", "<uuid>", ...],
//	  "sources": [      // #1285: plan, described, same order
//	    {"card_id": "<uuid>", "name": "Mountain", "zone": "battlefield", "tap": true},
//	    {"card_id": "<uuid>", "name": "Gold", "zone": "battlefield", "sacrifice": true},
//	    {"card_id": "<uuid>", "name": "Simian Spirit Guide", "zone": "hand", "exile": true}
//	  ],
//	  "missing": null,  // or ["{R}", "{1}"] when ok is false
//	  "cost": "{2}"     // the cost string this cast PAYS
//	}
//
// Errors: 400 on missing/malformed query params; 403 on a
// non-admin caller asking for someone else's preview; 404 when
// the game / card doesn't exist.
func autoTapPreview(c Config, w http.ResponseWriter, r *http.Request) error {
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
	if p.Role == auth.RoleSpectator || p.PlayerID == uuid.Nil {
		return httpError(http.StatusForbidden, "spectators have no auto-tap preview")
	}
	cardIDStr := r.URL.Query().Get("card")
	if cardIDStr == "" {
		return httpError(http.StatusBadRequest, "card query param is required")
	}
	cardID, err := uuid.Parse(cardIDStr)
	if err != nil {
		return httpError(http.StatusBadRequest, "card is not a valid UUID")
	}
	xValue := 0
	if xs := r.URL.Query().Get("x"); xs != "" {
		v, err := strconv.Atoi(xs)
		if err != nil || v < 0 {
			return httpError(http.StatusBadRequest, "x must be a non-negative integer")
		}
		xValue = v
	}
	// #916: ?phyrexian=<n> is the announce-time claim the cost
	// prompt is collecting — how many of the cost's Phyrexian
	// symbols are being paid with 2 life each (CR 107.4f). The
	// preview plans the MANA HALF ONLY, exactly as the engine's
	// auto-tapper does, so the stepper's readout answers "what will
	// this still cost me in mana" rather than tapping a land for a
	// pip the player just said they would pay with life.
	phyrexian := 0
	if ps := r.URL.Query().Get("phyrexian"); ps != "" {
		v, err := strconv.Atoi(ps)
		if err != nil || v < 0 {
			return httpError(http.StatusBadRequest, "phyrexian must be a non-negative integer")
		}
		phyrexian = v
	}
	locked, err := uuidListParam(r.URL.Query().Get("exclude"), "exclude")
	if err != nil {
		return err
	}
	excluded := make(map[uuid.UUID]bool, len(locked))
	for _, eid := range locked {
		excluded[eid] = true
	}
	g, err := c.Lobby.LookupGame(id)
	if err != nil {
		return httpError(http.StatusNotFound, err.Error())
	}
	if g == nil {
		return httpError(http.StatusNotFound, "game not found")
	}
	// Look the card up via the game's effect-API surface so the
	// preview consumes the same ManaCost the cost validator
	// would. The ParseCost gate already accepts an empty
	// ManaCost as costless, so a not-found-card path returns a
	// clean 404 rather than a misleading "ok with empty plan."
	card, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return httpError(http.StatusNotFound, "card not found in game")
	}
	// ?ability=<index> asks about a CR 602 activated ability's mana
	// component instead of the card's printed cost. The X picker
	// opens for an ability exactly as it does for a spell, and
	// without this it would price Helm of Obedience's "{X}" against
	// the {4} printed in the corner — a live readout that is wrong
	// in both directions is worse than none.
	//
	// #1405: priced by g.PriceActivation, which calls the function
	// ActivateCatalogAbility pays (AbilityManaCostForTargetsForEffect)
	// — the board's activation modifiers (#1184, Boom Scholar), the
	// ability's own clause (#1296, the channel lands) and, when the
	// request names them, the target-dependent price (Dragonfire
	// Blade). No commander tax: an ability is not a cast. This branch
	// used to parse the printed cost itself, on the claim that the
	// engine applied no modifier to an ability — untrue since #1184.
	if as := r.URL.Query().Get("ability"); as != "" {
		idx, err := strconv.Atoi(as)
		if err != nil || idx < 0 {
			return httpError(http.StatusBadRequest, "ability must be a non-negative integer")
		}
		// ?targets= is optional. The X and Phyrexian pickers open
		// before targeting (CR 601.2b before 601.2c), so they send
		// none and read the no-target price — for a target-keyed
		// reduction, the printed cost, which is what the activation
		// charges when it names no target.
		targets, err := previewTargetsParam(r.URL.Query().Get("targets"))
		if err != nil {
			return err
		}
		params, err := abilityParamsFromPreviewQuery(r, xValue)
		if err != nil {
			return err
		}
		price, err := g.PriceActivation(p.PlayerID, cardID, idx, targets)
		switch {
		case errors.Is(err, game.ErrCardNotFound):
			return httpError(http.StatusNotFound, "card not found in game")
		case errors.Is(err, game.ErrInvalidParam):
			return httpError(http.StatusNotFound, "ability index out of range")
		case err != nil:
			return httpError(http.StatusBadRequest, "this activation cannot be priced: "+err.Error())
		}
		// #1422: what the activation has already SPENT is not the
		// plan's to spend again — the source itself when the cost
		// taps or sacrifices it, the permanents and cards named to
		// its other components, and the waterbend taps. The same
		// function ActivateCatalogAbility excludes with, so the preview
		// can no longer plan a Castle Vantress's own {T} for the
		// {2}{U}{U} of "{2}{U}{U}, {T}: Scry 2".
		for eid := range game.ActivationAutoTapExclusions(cardID, price.Ability.Cost, params) {
			excluded[eid] = true
		}
		// #1422, CR 701.67a: each waterbend tap pays {1} instead of
		// mana — subtracted after pricing and before the Phyrexian
		// strike, the activation's own order, by the activation's own
		// function.
		total := game.WaterbendReduced(price.Total, xValue, len(params.WaterbendIDs))
		spend := game.ManaSpendForAbility(price.Source)
		cost := strikePhyrexianForPreview(g, p.PlayerID, total, spend, phyrexian)
		// #1212: no source wish for an ability. "If mana from a
		// Treasure was spent to activate this ability" (Forsworn
		// Paladin, Jetmir's Fixer) is real printed text and is
		// declared out of scope there — the record will carry the
		// kinds, nothing reads them off an activation yet, and the
		// catalog declaration is per CARD rather than per ability.
		//
		// `cost` stays the printed string, as on the cast branch: the
		// modifiers are generic, and plan / missing carry the total.
		return writeAutoTapPreview(g, p.PlayerID, cost, xValue, excluded,
			price.Ability.Cost.Mana, spend, 0, w)
	}
	// #696: the whole of the cast's price, from the engine's one
	// pricer. The alternative cost claimed at announce, a granted
	// permission's flat override (airbend's {2}, cascade's {0}), the
	// "spend mana as though any colour" fold, the commander tax, the
	// mana half of the announced optional costs, the board's cost
	// modifiers and the convoke/waterbend subtraction all land in one
	// call, priced for the face and the source zone the cast will
	// actually name.
	//
	// ADR 0048 addendum §15: the modifier pass inside it also applies
	// the card's own self modifiers (affinity, Ghalta). The preview
	// still has no targets on its query string, so it prices with
	// none: a per-target surcharge (Fireball, strive) reads as the
	// one-target price, which is what the X picker shows alongside the
	// printed clause (CardView.target_cost_notes).
	params, err := castParamsFromPreviewQuery(r, xValue)
	if err != nil {
		return err
	}
	// #1242: and what the announcement has already SPENT is not the
	// plan's to spend again — the convoke / waterbend taps, and the
	// cards and permanents named to the additional cost. The same
	// list CastSpell's auto-tap excludes (game.CastAutoTapExclusions),
	// so the preview shows the payment the cast will make: without it
	// a preview could crack the very Eldrazi Spawn the cast offers to
	// Village Rites and read "ok" for a cast the engine then refuses.
	for eid := range game.CastAutoTapExclusions(params) {
		excluded[eid] = true
	}
	price, err := g.PriceCast(p.PlayerID, card, params)
	if err != nil {
		return httpError(http.StatusBadRequest, "this cast cannot be priced: "+err.Error())
	}
	// price.Card, not `card`: the face the cast announces is
	// materialised by the pricer, and the spend context is read off
	// the card type, which a modal DFC's two faces need not share
	// (ADR 0034).
	spend := game.ManaSpendForCast(price.Card)
	cost := strikePhyrexianForPreview(g, p.PlayerID, price.Total, spend, phyrexian)
	// The cost string reported back is the one this cast PAYS, not
	// the one in the card's corner — a flashed-back Think Twice reads
	// "{2}{U}" and an airbent permanent reads "{2}". The commander tax
	// and the cost modifiers are not folded into the string (they are
	// generic and the string is Scryfall notation); `plan` and
	// `missing` are the authority on the total, and they are priced
	// with both.
	// #1212: the spell's own source wish, so the plan this preview
	// hands the client is the plan the engine's own auto-tapper would
	// have made. The preview's plan is what the client actually taps;
	// without this a Hired Hexblade previewed and tapped through the
	// UI would be paid off a Sol Ring and draw nothing, while the
	// same cast with AutoTap set would be paid off the Treasure and
	// draw. Two routes, one answer.
	return writeAutoTapPreview(g, p.PlayerID, cost, xValue, excluded,
		price.Paid, spend, game.WantedManaSourcesFor(price.Card), w)
}

// castParamsFromPreviewQuery reads the announce-time half of the cast
// off the preview's query string (#696). Everything here changes the
// PRICE, which is why the endpoint takes it at all: the same values
// the client will put in the cast_spell payload it is previewing.
//
// Malformed input is a 400 rather than a silent default, for the
// reason the endpoint exists — a preview that quietly priced a
// different cast from the one the button will send is worse than no
// preview.
func castParamsFromPreviewQuery(r *http.Request, xValue int) (game.CastSpellParams, error) {
	q := r.URL.Query()
	params := game.CastSpellParams{
		FromZone:        q.Get("from_zone"),
		AlternativeCost: q.Get("alternative_cost"),
		XValue:          xValue,
	}
	if fs := q.Get("face"); fs != "" {
		v, err := strconv.Atoi(fs)
		if err != nil || v < 0 {
			return params, httpError(http.StatusBadRequest, "face must be a non-negative integer")
		}
		params.Face = v
	}
	if os := q.Get("optional_costs"); os != "" {
		for _, raw := range strings.Split(os, ",") {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			v, err := strconv.Atoi(raw)
			if err != nil || v < 0 {
				return params, httpError(http.StatusBadRequest, "optional_costs must be a comma-separated list of non-negative integers")
			}
			params.OptionalCosts = append(params.OptionalCosts, v)
		}
	}
	ids, err := uuidListParam(q.Get("tap_ids"), "tap_ids")
	if err != nil {
		return params, err
	}
	params.TapIDs = ids
	// #1242: the additional cost's named payments. They do not change
	// the price; they change what the auto-tapper may spend on it.
	if params.SacrificeIDs, err = uuidListParam(q.Get("sacrifice_ids"), "sacrifice_ids"); err != nil {
		return params, err
	}
	if params.DiscardIDs, err = uuidListParam(q.Get("discard_ids"), "discard_ids"); err != nil {
		return params, err
	}
	return params, nil
}

// abilityParamsFromPreviewQuery reads the ?ability= branch's named
// cost payments off the query string (#1422), under the field names
// the activate_ability payload uses for them. None changes the price
// except waterbend_ids, which pays {1} per tap; all of them change
// what the auto-tapper may spend on the mana half, exactly as they do
// in ActivateCatalogAbility.
func abilityParamsFromPreviewQuery(r *http.Request, xValue int) (game.ActivateAbilityParams, error) {
	q := r.URL.Query()
	params := game.ActivateAbilityParams{XValue: xValue}
	for _, f := range []struct {
		name string
		dst  *[]uuid.UUID
	}{
		{"tap_ids", &params.TapIDs},
		{"sacrifice_ids", &params.SacrificeIDs},
		{"discard_ids", &params.DiscardIDs},
		{"exile_ids", &params.ExileIDs},
		{"waterbend_ids", &params.WaterbendIDs},
	} {
		ids, err := uuidListParam(q.Get(f.name), f.name)
		if err != nil {
			return params, err
		}
		*f.dst = ids
	}
	return params, nil
}

// uuidListParam parses a comma-separated UUID query param, naming
// itself in the 400 so the caller can tell which list was malformed.
func uuidListParam(raw, name string) ([]uuid.UUID, error) {
	if raw == "" {
		return nil, nil
	}
	var out []uuid.UUID
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, httpError(http.StatusBadRequest, name+" must be a comma-separated UUID list")
		}
		out = append(out, id)
	}
	return out, nil
}

// previewTargetsParam parses the activation preview's ?targets= list
// (#1405): comma-separated `<kind>:<uuid>` entries, kind `card` or
// `player` — the two kinds a target-reading cost can look at. A
// TargetRef's Slot and Mode are not carried: no cost modifier reads
// them, and the preview asks only for the price.
func previewTargetsParam(raw string) ([]game.TargetRef, error) {
	if raw == "" {
		return nil, nil
	}
	const malformed = "targets must be a comma-separated list of card:<uuid> or player:<uuid>"
	var out []game.TargetRef
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		kind, idStr, ok := strings.Cut(entry, ":")
		ref := game.TargetRef{Kind: game.TargetRefKind(kind)}
		if !ok || (ref.Kind != game.TargetCard && ref.Kind != game.TargetPlayer) {
			return nil, httpError(http.StatusBadRequest, malformed)
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, httpError(http.StatusBadRequest, malformed)
		}
		ref.ID = id
		out = append(out, ref)
	}
	return out, nil
}

// strikePhyrexianForPreview removes the Phyrexian symbols the caller
// says they are paying with life, so the preview plans only the mana
// the announcement still owes (#916).
//
// game.PhyrexianLifePlan is the SAME strike the engine makes, reading
// the same pool, so the preview and the payment pick the same
// symbols. It clamps rather than rejects: the preview is advisory,
// and an over-claim is the announce gate's refusal to make
// (strikePhyrexianLifeLocked, CR 601.2b / CR 119.4), not a reason to
// answer a read-only question with a 400.
func strikePhyrexianForPreview(
	g *game.Game,
	playerID uuid.UUID,
	cost game.ParsedCost,
	spend game.ManaSpendContext,
	claimed int,
) game.ParsedCost {
	if claimed <= 0 {
		return cost
	}
	if have := cost.PhyrexianSymbols(); claimed > have {
		claimed = have
	}
	seat := g.PlayerByIDForEffect(playerID)
	if seat == nil {
		return cost
	}
	reduced, _ := game.PhyrexianLifePlan(cost, seat.ManaPool, spend, claimed)
	return reduced
}

// writeAutoTapPreview renders the auto-tap preview body for an
// already-resolved cost. Shared by the cast branch and the
// activated-ability branch so the two can never disagree about the
// response shape.
func writeAutoTapPreview(
	g *game.Game,
	playerID uuid.UUID,
	cost game.ParsedCost,
	xValue int,
	excluded map[uuid.UUID]bool,
	costStr string,
	spend game.ManaSpendContext,
	prefer game.ManaSourceKinds,
	w http.ResponseWriter,
) error {
	plan, ok := g.AutoTapPlanPreferringExcluding(playerID, cost, xValue, excluded, prefer)
	// #1285: `sources` describes each planned source — where it is and
	// what paying with it costs — beside the bare `plan` ID list, which
	// is unchanged for every reader that only wanted the IDs. A plan
	// entry stopped meaning "an untapped permanent" twice over: #1228
	// can plan a Spirit Guide out of the HAND, and #1242 can crack a
	// Gold or an Eldrazi Spawn without tapping it. A client that looked
	// the IDs up on the battlefield showed the first as nothing at all.
	type source struct {
		CardID     string `json:"card_id"`
		Name       string `json:"name,omitempty"`
		Zone       string `json:"zone,omitempty"`
		Tap        bool   `json:"tap,omitempty"`
		Sacrifice  bool   `json:"sacrifice,omitempty"`
		ExileCards bool   `json:"exile,omitempty"`
	}
	type response struct {
		OK      bool     `json:"ok"`
		Plan    []string `json:"plan,omitempty"`
		Sources []source `json:"sources,omitempty"`
		Missing []string `json:"missing,omitempty"`
		Cost    string   `json:"cost"`
	}
	body := response{OK: ok, Cost: costStr}
	if ok {
		body.Plan = make([]string, len(plan))
		body.Sources = make([]source, len(plan))
		for i, e := range plan {
			body.Plan[i] = e.CardID.String()
			body.Sources[i] = source{
				CardID:     e.CardID.String(),
				Name:       e.Name,
				Zone:       string(e.Zone),
				Tap:        e.Taps,
				Sacrifice:  e.Sacrifices,
				ExileCards: e.Exiles,
			}
		}
	} else {
		// On miss, surface the unpaid symbols so the client UI can
		// reuse the same "missing {R}{R}" copy the strict-mode
		// override toast renders.
		seat := g.PlayerByIDForEffect(playerID)
		if seat != nil {
			// #352: the breakdown is computed under the same spend
			// context the cast will pay under, so a pool of Ancient
			// Ziggurat mana does not report "missing nothing" for a
			// spell it cannot legally fund.
			body.Missing = seat.ManaPool.MissingFor(cost, xValue, spend)
		}
	}
	return writeJSON(w, http.StatusOK, body)
}

// downloadReplay streams the per-game JSONL replay log back to the
// caller. Each line is one protocol.SnapshotPayload (the same shape
// the WS layer broadcasts), in the order Apply produced them. A
// downstream consumer can deserialize line-by-line to reconstruct
// the full game timeline.
//
// The replay log on disk holds the UNFILTERED view — every seat's
// hand and full library order, for every snapshot. Access is
// therefore governed on two independent axes, WHEN and WHAT:
//
// WHEN. The log is live hidden information for as long as the game
// runs. Admins may download at any time; players and spectators
// bound to this game only once the game has ended (by then every
// snapshot is post-game history, not an in-progress scouting feed).
//
// WHAT. "The game is over" is not the same as "everyone may now read
// everyone's hidden information." Admins get the log verbatim — bug
// triage and the pinned bug-report replay need full fidelity. Every
// other caller gets it streamed through protocol.FilterViewFor for
// their own seat, which makes the endpoint consistent with the live
// WS path: ws.Hub already filters every snapshot it broadcasts with
// exactly this function, so an unfiltered replay was the one way to
// obtain state the socket had deliberately withheld. After this, a
// replay can never show a viewer more than they legitimately saw at
// the table — and the WHEN gate above becomes defence in depth
// rather than the only thing standing between a seated player and
// the whole pod's hidden information.
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
	if p.Role != auth.RoleAdmin && room.Game.CurrentState() != game.StateEnded {
		return httpError(http.StatusForbidden, "replay is available once the game has ended")
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
	f, err := os.Open(path)
	if err != nil {
		return httpError(http.StatusInternalServerError, "open replay: "+err.Error())
	}
	// ServeContent does NOT close its ReadSeeker, and it completes
	// before this handler returns — defer is the whole cleanup story.
	defer func() { _ = f.Close() }()
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set(
		"Content-Disposition",
		`attachment; filename="`+id.String()+`.jsonl"`,
	)
	if p.Role == auth.RoleAdmin {
		// Verbatim. ServeContent is a zero-copy sendfile path and
		// brings Range support plus a Content-Length along with it —
		// worth keeping on the one path that doesn't transform.
		http.ServeContent(w, r, id.String()+".jsonl", info.ModTime(), f)
		return nil
	}
	streamFilteredReplay(c, w, f, replayViewerID(p))
	return nil
}

// replayViewerID converts a principal into the viewerID string
// protocol.FilterViewFor expects. A seated player is their own UUID;
// anyone else is "", the spectator view. The uuid.Nil coercion
// matters — Nil stringifies to the all-zero UUID, which matches no
// seat and is NOT what FilterViewFor documents as "no seat". Mirrors
// ws.viewerIDForFilter, which is unexported in that package.
func replayViewerID(p auth.Principal) string {
	if p.Role != auth.RolePlayer || p.PlayerID == uuid.Nil {
		return ""
	}
	return p.PlayerID.String()
}

// streamFilteredReplay copies the JSONL replay from src to w, passing
// each snapshot's GameView through FilterViewFor for the given viewer
// and preserving line order and Seq numbers.
//
// Decoding with json.Decoder rather than bufio.Scanner is deliberate:
// one record is an entire GameView and routinely exceeds Scanner's
// 64 KiB default token size, which would truncate the replay
// mid-stream and report no error at all. Encoder.Encode appends the
// newline, so the output is JSONL by construction.
//
// Errors after the first byte cannot become an HTTP status — the
// header is already committed — so a torn tail (a partial append from
// a crash) ends the stream cleanly, leaving the client a shorter but
// well-formed JSONL document rather than a corrupt one.
func streamFilteredReplay(c Config, w http.ResponseWriter, src io.Reader, viewerID string) {
	dec := json.NewDecoder(src)
	enc := json.NewEncoder(w)
	for n := 0; ; n++ {
		var payload protocol.SnapshotPayload
		if err := dec.Decode(&payload); err != nil {
			if !errors.Is(err, io.EOF) {
				logReplayWarning(c, "replay stream ended early", err, n)
			}
			return
		}
		payload.Game = protocol.FilterViewFor(payload.Game, viewerID)
		if err := enc.Encode(payload); err != nil {
			// Client hung up mid-download. Nothing to report.
			return
		}
	}
}

func logReplayWarning(c Config, what string, err error, line int) {
	if c.Log == nil {
		return
	}
	c.Log.Warn("replay: "+what, "err", err, "line", line)
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
// Exactly one of Deck (a pre-built deck ID) and Source (a decklist)
// must be set; Format is optional and only narrows how Source is
// parsed when the caller can't match the file extension to a format
// string. The server echoes back a parsed + validated summary on
// success.
type uploadDeckRequest struct {
	// Format is one of "text", "moxfield", or empty (auto-detect
	// from the first non-whitespace byte: '{' → moxfield, else text).
	Format string `json:"format,omitempty"`
	// Source is the raw decklist payload. For "text" format, the
	// plain-text decklist. For "moxfield", the JSON export bytes.
	Source string `json:"source"`
	// Deck is a pre-built deck ID from GET /decks — the path for a
	// player who wants to play rather than to bring a list. Exactly
	// one of Deck and Source may be set.
	//
	// It is a field on THIS request and not a route of its own
	// because the deck it names is installed by the same
	// ParseText → Resolve → Validate → SetDeck pipeline a pasted
	// list is: internal/decks hands over decklist TEXT, not resolved
	// cards, precisely so there is one legality path in the server
	// and a pre-built deck cannot be legal by a rule an uploaded one
	// is not held to. POST /games/{id}/seats/bot has taken the same
	// shape since S31 for the same reason.
	Deck string `json:"deck,omitempty"`
	// PlayerID is the seat this deck is for. A RolePlayer caller can
	// only set their own deck — we cross-check against the principal
	// below.
	PlayerID uuid.UUID `json:"player_id"`
}

// uploadDeckResponse describes the accepted deck. Mirrors the lobby
// seat update the caller will see via GET /games/{id}, but returned
// inline so the client doesn't have to refetch.
type uploadDeckResponse struct {
	Game      GameMeta `json:"game"`
	DeckName  string   `json:"deck_name"`
	CardCount int      `json:"card_count"`
	// DeckID echoes the pre-built deck that was installed on
	// POST /games/{id}/decks (empty for an uploaded list), or the
	// library deck id on POST /games/{id}/decks/{deck_id} — see
	// seatLibraryDeck. The client uses it to confirm the seat is
	// holding the deck the player picked rather than guessing from the
	// name.
	DeckID     string   `json:"deck_id,omitempty"`
	Commanders []string `json:"commanders"`
	// Warnings is a non-fatal violation list (e.g. sideboard ignored).
	// The accepted deck is already installed when warnings is
	// non-empty; treat it as advisory.
	Warnings []deck.Violation `json:"warnings,omitempty"`
	// Unimplemented names the accepted deck's cards that print rules
	// the engine will not carry out — see game.Unimplemented. Not a
	// violation and not a warning: the deck is legal and the game
	// will run, those cards just behave as manual sandbox cards and
	// the player moves the pieces themselves.
	//
	// Deck upload is the best moment there is to say so. It is a
	// single honest sentence about a hundred cards, read once,
	// before anyone has formed an expectation — as against the
	// alternative, which is what actually happened on 2026-09-10:
	// five separate bug reports (#321, #324, #325, #332, #333) from
	// five separate mid-game surprises.
	//
	// Distinct names in decklist order, deduped — a deck with four
	// Lightning Bolts wants to hear about Lightning Bolt once.
	Unimplemented []string `json:"unimplemented,omitempty"`
}

// detectDeckFormat guesses a decklist's format from its first non-
// whitespace bytes, for a caller that left Format empty. URLs are
// detected first since they're unambiguous ("http://" or "https://"
// prefix); JSON next (leading `{`); everything else is plain text.
//
// Shared by resolveDeckSource, so parsing agrees with its own
// default, and by uploadDeck, which needs to know the format that was
// actually used to decide whether — and as what — a signed-in
// player's paste is worth saving to their deck library (ADR 0051
// decision 7, S34 sub-PR 5).
func detectDeckFormat(source string) string {
	trimmed := strings.TrimLeft(source, " \t\r\n")
	switch {
	case strings.HasPrefix(trimmed, "http://"), strings.HasPrefix(trimmed, "https://"):
		return "url"
	case strings.HasPrefix(trimmed, "{"):
		return "moxfield"
	default:
		return "text"
	}
}

// resolveDeckSource is the shared parse → resolve → validate pipeline
// behind POST /games/{id}/decks and POST /games/{id}/seats/bot. It
// returns the installed-ready list and any non-fatal warnings. When
// `written` is true the handler has already answered the request (a
// 422 with the violation list) and the caller must return nil.
func resolveDeckSource(ctx context.Context, c Config, w http.ResponseWriter, format, source string) (*deck.List, []deck.Violation, bool, error) {
	if format == "" {
		format = detectDeckFormat(source)
	}

	var (
		deckName string
		entries  []deck.Entry
		perr     error
	)
	switch format {
	case "text":
		entries, perr = deck.ParseText(source)
	case "moxfield":
		deckName, entries, perr = deck.ParseMoxfield([]byte(source))
	case "url":
		client := c.DeckHTTPClient
		if client == nil {
			client = deck.DefaultClient()
		}
		deckName, entries, perr = deck.FetchFromURL(ctx, client, strings.TrimSpace(source))
		if perr != nil {
			// Fetcher failures (unknown host, private deck, upstream
			// down, etc.) surface via FetchViolation as typed 422
			// entries so the client renders them the same way it
			// renders validation violations.
			if v, ok := deck.FetchViolation(strings.TrimSpace(source), perr); ok {
				return nil, nil, true, writeDeckViolations(w, perr.Error(), []deck.Violation{v}, nil)
			}
			// Unknown-mechanic inside the fetched payload (e.g. a
			// Moxfield deck with a companion slot) bubbles through
			// here; fall into the existing Resolve-failure path
			// below by leaving perr set.
		}
	default:
		return nil, nil, false, httpError(http.StatusBadRequest, fmt.Sprintf("unknown deck format %q", format))
	}
	if perr != nil {
		return nil, nil, false, httpError(http.StatusBadRequest, perr.Error())
	}

	list, err := deck.Resolve(c.Cards, deckName, entries)
	if err != nil {
		// Resolve failures that carry a per-card Violation list are
		// surfaced with the same 422 `{"error", "violations"}` shape
		// the validator uses, so the client has one schema to handle.
		var uce *deck.UnknownCardError
		if errors.As(err, &uce) {
			return nil, nil, true, writeDeckViolations(w, err.Error(), uce.Violations(), nil)
		}
		var ume *deck.UnsupportedMechanicError
		if errors.As(err, &ume) {
			return nil, nil, true, writeDeckViolations(w, err.Error(), ume.Violations(), nil)
		}
		return nil, nil, false, httpError(http.StatusBadRequest, err.Error())
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
				// The non-fatal classes: the deck imports and plays,
				// with something declared. ADR 0034 added the second
				// — a transform or split card is imported as its
				// front half rather than refused, and the banner
				// says so. Rejecting a whole deck over a card that
				// is merely cosmetically simplified is the wrong
				// trade; saying nothing is what produced #265.
				if v.Code == deck.CodeSideboardUnsupported ||
					v.Code == deck.CodeUnsupportedLayout {
					warnings = append(warnings, v)
					continue
				}
				fatal = append(fatal, v)
			}
			if len(fatal) > 0 {
				return nil, nil, true, writeDeckViolations(w, "deck has validation errors", fatal, warnings)
			}
		} else {
			return nil, nil, false, verr
		}
	}

	return list, warnings, false, nil
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
	format, source, deckID, err := uploadDeckChoice(body)
	if err != nil {
		return err
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

	list, warnings, written, err := resolveDeckSource(r.Context(), c, w, format, source)
	if written || err != nil {
		return err
	}
	// A pre-built deck's decklist TEXT carries no name — plain-text
	// format has nowhere to put one, which is why an uploaded .txt is
	// nameless too. The deck has a name, though, and the seat is about
	// to be labelled with whatever goes in here: without this the
	// lobby shows "deck ready" where it could say "Deep Roots".
	if deckID != "" && list.Name == "" {
		list.Name = prebuiltDeckName(deckID)
	}

	gameCards := list.ToGameCards()
	meta, err := c.Lobby.SetDeck(id, body.PlayerID, list.Name, gameCards)
	if err != nil {
		return err
	}

	commanders := make([]string, 0, len(list.Commanders))
	for _, cc := range list.Commanders {
		commanders = append(commanders, cc.Name)
	}

	// Save to the caller's deck library (ADR 0051 decision 7, S34
	// sub-PR 5) — see saveToLibrary for exactly when this does
	// something and the seats.deck_id it leaves behind.
	libraryDeckID := saveToLibrary(r.Context(), c, p, deckID, format, source, list, commanders)
	if serr := c.Lobby.SetSeatDeckID(id, body.PlayerID, libraryDeckID); serr != nil {
		logDeckLibraryWarning(c, "set seat deck_id failed", serr, "game_id", id, "player_id", body.PlayerID)
	}

	return writeJSON(w, http.StatusOK, uploadDeckResponse{
		Game:          meta,
		DeckName:      list.Name,
		CardCount:     len(list.Commanders) + len(list.Mainboard),
		DeckID:        deckID,
		Commanders:    commanders,
		Warnings:      warnings,
		Unimplemented: game.UnimplementedNames(gameCards),
	})
}

// saveToLibrary creates or updates the caller's deck-library row for
// this upload (ADR 0051 decision 7, S34 sub-PR 5), and returns the
// library deck id the seat should now be linked to — "" when nothing
// was saved.
//
// Only three things gate a save, all deliberate:
//
//   - deckID must be "" — a pre-built catalog pick (uploadDeckRequest.
//     Deck) has its own id system and is never a library row; the
//     "source" it hands resolveDeckSource is the catalog's TEXT, not
//     anything the player pasted.
//   - p.UserID must be non-zero — guests have nowhere to own a row.
//   - resolvedFormat must be "moxfield" or "text", matching
//     decks.source_format. A "url" request's source is a link, not
//     the decklist text decision 7 means by "what the player pasted";
//     re-seating from it would mean a network call at seat time
//     rather than a re-parse of stored text, which is not what a
//     library is for.
//
// A save failure is logged and does not fail the request: SetDeck has
// already installed the deck on the seat by the time this runs, and
// telling the player their upload failed when it didn't would be
// worse than a library row they can save again by re-uploading.
//
// The update rule (documented on decklibrary.Store.Upsert and in
// docs/lobby.md): a row already owned by this caller with the same
// name is updated in place; anything else inserts a new one. A
// plain-text paste carries no name of its own (deck.ParseText has
// nowhere to put one), so an empty list.Name falls back to the first
// commander's name rather than saving a blank row every time.
func saveToLibrary(ctx context.Context, c Config, p auth.Principal, deckID, format, source string, list *deck.List, commanders []string) string {
	if deckID != "" || p.UserID == uuid.Nil || c.DeckLibrary == nil {
		return ""
	}
	resolvedFormat := format
	if resolvedFormat == "" {
		resolvedFormat = detectDeckFormat(source)
	}
	if resolvedFormat != "text" && resolvedFormat != "moxfield" {
		return ""
	}
	name := list.Name
	if name == "" {
		name = libraryFallbackName(list)
	}
	cardCount := len(list.Commanders) + len(list.Mainboard)
	saved, err := c.DeckLibrary.Upsert(ctx, p.UserID, name, resolvedFormat, source, commanders, cardCount)
	if err != nil {
		logDeckLibraryWarning(c, "save deck to library failed", err, "user_id", p.UserID)
		return ""
	}
	return saved.ID.String()
}

// libraryFallbackName names a deck being saved to the library when
// its parsed List carries no name — every plain-text paste, since
// that format has nowhere to put one (deck.ParseText). The first
// commander is a more useful label than a blank row. Two different
// decks on the same commander and no other name collide under the
// update rule (same owner, same name) exactly as two Moxfield exports
// named identically would; a player who wants both kept separate
// names one of them.
func libraryFallbackName(list *deck.List) string {
	if len(list.Commanders) > 0 {
		return list.Commanders[0].Name
	}
	return "Untitled deck"
}

// logDeckLibraryWarning logs a non-fatal deck-library failure (see
// saveToLibrary). Mirrors logReplayWarning: c.Log is nil in tests that
// don't wire one, and a swallowed warning there is fine — nothing in
// the request path depends on it.
func logDeckLibraryWarning(c Config, what string, err error, args ...any) {
	if c.Log == nil {
		return
	}
	c.Log.Warn("decklibrary: "+what, append([]any{"err", err}, args...)...)
}

// seatLibraryDeck handles POST /games/{id}/decks/{deck_id}: seat a
// deck already saved to the caller's library (ADR 0051 decision 7,
// S34 sub-PR 5), without re-pasting it.
//
// Only a RolePlayer session already seated in this game may call it,
// and only for their own seat — there is no player_id in the body,
// unlike POST /games/{id}/decks, because a RolePlayer session names
// exactly one seat and there is nothing to disambiguate. RoleAdmin is
// refused outright rather than allowed "any seat" the way it is on
// the paste-upload route: the authorization that matters here is deck
// ownership (below), and an admin session never owns a deck — it has
// no UserID (ADR 0051 decision 2) — so admin access to this route
// could never do anything but 403 one step later anyway.
//
// The stored source_text is re-parsed through the same
// parse → resolve → validate pipeline an upload takes
// (resolveDeckSource), against the catalog THIS server has loaded
// right now — never the catalog at save time — so a card that
// stopped resolving (a rename, a ban, a dump that dropped it) surfaces
// as the same 422 violation list an upload would give, rather than
// installing something that silently changed.
func seatLibraryDeck(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.Cards == nil || c.Cards.Count() == 0 {
		return httpError(http.StatusServiceUnavailable, "card index not loaded; run scripts/scryfall-refresh.sh")
	}
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	deckID, err := uuid.Parse(r.PathValue("deck_id"))
	if err != nil {
		return httpError(http.StatusBadRequest, "invalid deck id")
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.Role != auth.RolePlayer || p.GameID != id {
		return httpError(http.StatusForbidden, "not a seat in this game")
	}
	if c.DeckLibrary == nil {
		return httpError(http.StatusServiceUnavailable, "deck library not configured")
	}

	saved, err := c.DeckLibrary.Get(r.Context(), deckID)
	if errors.Is(err, decklibrary.ErrNotFound) {
		return httpError(http.StatusNotFound, "deck not found")
	}
	if err != nil {
		return err
	}
	if saved.OwnerID != p.UserID {
		return httpError(http.StatusForbidden, "not your deck")
	}

	list, warnings, written, err := resolveDeckSource(r.Context(), c, w, saved.SourceFormat, saved.SourceText)
	if written || err != nil {
		return err
	}
	// The library's display name is authoritative, regardless of
	// whether the re-parse recovers one of its own — a plain-text
	// source never carries one (deck.ParseText), which would otherwise
	// re-label a renamed library deck back to blank on every reseat.
	list.Name = saved.Name

	gameCards := list.ToGameCards()
	meta, err := c.Lobby.SetDeck(id, p.PlayerID, list.Name, gameCards)
	if err != nil {
		return err
	}
	if serr := c.Lobby.SetSeatDeckID(id, p.PlayerID, saved.ID.String()); serr != nil {
		logDeckLibraryWarning(c, "set seat deck_id failed", serr, "game_id", id, "player_id", p.PlayerID)
	}

	commanders := make([]string, 0, len(list.Commanders))
	for _, cc := range list.Commanders {
		commanders = append(commanders, cc.Name)
	}
	return writeJSON(w, http.StatusOK, uploadDeckResponse{
		Game:          meta,
		DeckName:      list.Name,
		CardCount:     len(list.Commanders) + len(list.Mainboard),
		DeckID:        saved.ID.String(),
		Commanders:    commanders,
		Warnings:      warnings,
		Unimplemented: game.UnimplementedNames(gameCards),
	})
}

// myDeckInfo is one entry in GET /me/decks.
type myDeckInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Commanders []string  `json:"commanders"`
	CardCount  int       `json:"card_count"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// myDecksResponse is the body of GET /me/decks.
type myDecksResponse struct {
	Decks []myDeckInfo `json:"decks"`
}

// myDecks handles GET /me/decks: the caller's saved decks (ADR 0051
// decision 7, S34 sub-PR 5), newest updated first.
//
// 401 for any principal without a UserID — a guest's RolePlayer
// session, an admin session, or an identified session minted by a
// deployment with no database — not only for a missing credential,
// which auth.Middleware already turns into a 401 on its own. There is
// nothing partial to show: a UserID-less principal owns no decks by
// construction (decklibrary.Store.Upsert requires one).
func myDecks(c Config, w http.ResponseWriter, r *http.Request) error {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.UserID == uuid.Nil {
		return httpError(http.StatusUnauthorized, "sign-in required")
	}
	library := c.DeckLibrary
	if library == nil {
		library = decklibrary.NoStore{}
	}
	decks, err := library.List(r.Context(), p.UserID)
	if err != nil {
		return err
	}
	out := make([]myDeckInfo, 0, len(decks))
	for _, d := range decks {
		commanders := d.Commanders
		if commanders == nil {
			commanders = []string{}
		}
		out = append(out, myDeckInfo{
			ID:         d.ID.String(),
			Name:       d.Name,
			Commanders: commanders,
			CardCount:  d.CardCount,
			UpdatedAt:  d.UpdatedAt,
		})
	}
	return writeJSON(w, http.StatusOK, myDecksResponse{Decks: out})
}

// addBotRequest is the request shape for POST /games/{id}/seats/bot.
// The deck travels exactly as it does for /decks — a decklist
// (text, Moxfield JSON or a URL) — and goes through the same parse,
// resolve and validate pipeline, so a bot cannot be seated with a
// deck a human couldn't upload.
type addBotRequest struct {
	// Tier is the policy tier; must be one GET /bot/options reports
	// as available on THIS server — which depends on how it is
	// configured, since the model tiers need a model endpoint. An
	// unavailable tier is a 422 rather than a silent downgrade.
	Tier string `json:"tier"`
	// Deck is a curated-deck ID from GET /bot/options. This is the
	// player-facing path: pick a tier and a deck and press add.
	Deck string `json:"deck,omitempty"`
	// Name is the seat's display name. Defaults to "Bot N".
	Name string `json:"name,omitempty"`
	// Format / Source are a raw decklist, exactly as for
	// uploadDeckRequest — the escape hatch for the test harness and
	// for trying a list that is not in the catalog. Exactly one of
	// Deck or Source must be set.
	Format string `json:"format,omitempty"`
	Source string `json:"source,omitempty"`
}

// addBotResponse is the accepted seat.
type addBotResponse struct {
	Game     GameMeta         `json:"game"`
	PlayerID uuid.UUID        `json:"player_id"`
	DeckName string           `json:"deck_name"`
	Warnings []deck.Violation `json:"warnings,omitempty"`
	// Unimplemented is the same disclosure uploadDeckResponse carries,
	// for the same reason — see its doc comment. It was missing here,
	// which made the bot path the one place a deck could be installed
	// without anyone being told which of its cards the engine will not
	// carry out.
	//
	// Always empty for a curated deck: decks_test.go fails the build
	// if a card in one stops resolving to a registered Spec. It is the
	// `source` escape hatch that needs it, and that is exactly the
	// path issue #89 filed under "catalog-gap tolerance" — a deck with
	// an unrecognised card must not crash the bot, and the seat that
	// took it should say so rather than let the bot spend mana on
	// blanks in silence.
	Unimplemented []string `json:"unimplemented,omitempty"`
}

// botDeckSource turns the request's deck choice into a (format,
// source, deckID) triple for resolveDeckSource. Exactly one of
// `deck` and `source` may be set.
func botDeckSource(c Config, body addBotRequest) (format, source, deckID string, err error) {
	named := strings.TrimSpace(body.Deck)
	raw := strings.TrimSpace(body.Source)
	switch {
	case named != "" && raw != "":
		return "", "", "", httpError(http.StatusBadRequest, "send either deck or source, not both")
	case named != "":
		if c.BotDecks == nil {
			return "", "", "", httpError(http.StatusServiceUnavailable, "no bot deck catalog is configured on this server")
		}
		info, list, ok := c.BotDecks.Decklist(named)
		if !ok {
			return "", "", "", httpError(http.StatusUnprocessableEntity, fmt.Sprintf("unknown bot deck %q", named))
		}
		return "text", list, info.ID, nil
	case raw != "":
		return body.Format, body.Source, "", nil
	default:
		return "", "", "", httpError(http.StatusBadRequest, "deck or source is required")
	}
}

// botSeatAuthorised is the shared gate for the bot-seat routes:
// admin, or a player SEATED at this table. Anyone at the table may
// add or remove a bot while it is unstarted — ADR 0033 §9 is explicit
// that this is not admin-only, because the request was "add a bot to
// any unstarted table".
//
// A spectator's session also carries this GameID, so the seated check
// is `RolePlayer` with a real PlayerID and not merely a matching
// GameID. Adding a bot mutates the table; watching does not earn it.
// Mirrors the gate on autoTapPreview.
func botSeatAuthorised(r *http.Request, id uuid.UUID) error {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.Role == auth.RoleAdmin {
		return nil
	}
	if p.GameID != id {
		return httpError(http.StatusForbidden, "not a seat in this game")
	}
	if p.Role != auth.RolePlayer || p.PlayerID == uuid.Nil {
		return httpError(http.StatusForbidden, "only a seated player may change bot seats")
	}
	return nil
}

// botOptionsResponse is what the Add-bot picker renders from.
type botOptionsResponse struct {
	// Tiers is every declared tier, available or not, in picker
	// order. An unavailable tier is listed with available:false so
	// the UI can say what is coming instead of pretending the
	// difficulty slider has one notch.
	Tiers []aiseat.TierInfo `json:"tiers"`
	// Decks is the curated catalog. Empty when no deck source is
	// configured — the picker then falls back to pasting a decklist.
	Decks []aiseat.DeckInfo `json:"decks"`
	// Enabled is false when this server has no bot host at all, in
	// which case the client hides the Add-bot control rather than
	// offering a button that 503s.
	Enabled bool `json:"enabled"`
}

// botOptions handles GET /bot/options. Added in S31 sub-PR 4.
func botOptions(c Config, w http.ResponseWriter, _ *http.Request) error {
	out := botOptionsResponse{Tiers: aiseat.Tiers(), Decks: []aiseat.DeckInfo{}, Enabled: c.Bots != nil}
	if c.Bots != nil {
		// The host is the authority on what can actually play, so
		// reconcile the declared catalog against it rather than
		// trusting the table twice.
		offered := make(map[string]bool, len(c.Bots.Tiers()))
		for _, t := range c.Bots.Tiers() {
			offered[t] = true
		}
		// …and on why, when it can say. Availability alone tells a
		// player the tier is off; the reason tells whoever runs the
		// server how to turn it on.
		reasons, _ := c.Bots.(BotTierReasons)
		for i := range out.Tiers {
			out.Tiers[i].Available = offered[string(out.Tiers[i].Tier)]
			out.Tiers[i].Reason = ""
			if !out.Tiers[i].Available && reasons != nil {
				out.Tiers[i].Reason = reasons.TierReason(string(out.Tiers[i].Tier))
			}
		}
	} else {
		for i := range out.Tiers {
			out.Tiers[i].Available = false
			out.Tiers[i].Reason = "this server has no bot host configured"
		}
	}
	if c.BotDecks != nil {
		if decks := c.BotDecks.List(); len(decks) > 0 {
			out.Decks = decks
		}
	}
	return writeJSON(w, http.StatusOK, out)
}

// addBot handles POST /games/{id}/seats/bot. Added in S31 sub-PR 4.
func addBot(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.Bots == nil {
		return httpError(http.StatusServiceUnavailable, "bot seats are not enabled on this server")
	}
	if c.Cards == nil || c.Cards.Count() == 0 {
		return httpError(http.StatusServiceUnavailable, "card index not loaded; run scripts/scryfall-refresh.sh")
	}
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	if err := botSeatAuthorised(r, id); err != nil {
		return err
	}
	if r.Body == nil {
		return httpError(http.StatusBadRequest, "missing body")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxDeckBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var body addBotRequest
	if err := dec.Decode(&body); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return httpError(http.StatusRequestEntityTooLarge, fmt.Sprintf("deck source exceeds %d-byte limit", maxDeckBodyBytes))
		}
		return httpError(http.StatusBadRequest, fmt.Sprintf("invalid body: %s", err.Error()))
	}
	format, source, deckID, err := botDeckSource(c, body)
	if err != nil {
		return err
	}
	tier := strings.ToLower(strings.TrimSpace(body.Tier))
	known := false
	for _, t := range c.Bots.Tiers() {
		if t == tier {
			known = true
			break
		}
	}
	if !known {
		return fmt.Errorf("%w: %q (available: %s)", ErrUnknownBotTier, body.Tier, strings.Join(c.Bots.Tiers(), ", "))
	}

	list, warnings, written, err := resolveDeckSource(r.Context(), c, w, format, source)
	if written || err != nil {
		return err
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		meta, gerr := c.Lobby.Get(id)
		if gerr != nil {
			return gerr
		}
		bots := 0
		for _, s := range meta.Players {
			if s.IsBot {
				bots++
			}
		}
		name = fmt.Sprintf("Bot %d", bots+1)
	}
	gameCards := list.ToGameCards()
	meta, playerID, err := c.Lobby.AddBot(id, name, tier, deckID, list.Name, gameCards)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusCreated, addBotResponse{
		Game:          meta,
		PlayerID:      playerID,
		DeckName:      list.Name,
		Warnings:      warnings,
		Unimplemented: game.UnimplementedNames(gameCards),
	})
}

// removeBot handles DELETE /games/{id}/seats/bot/{player}. Added in
// S31 sub-PR 4.
func removeBot(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	if err := botSeatAuthorised(r, id); err != nil {
		return err
	}
	playerID, err := uuid.Parse(r.PathValue("player"))
	if err != nil {
		return httpError(http.StatusBadRequest, "invalid player id")
	}
	meta, err := c.Lobby.RemoveBot(id, playerID)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, meta)
}

// logout asks the authenticator to revoke the caller's credential and
// clears the session cookie. Unauthenticated — we want a client with
// an already-expired token to be able to reach this endpoint to flush
// its cookie without hitting a 401 first. Always returns 204 so the
// client can safely treat the response as idempotent.
//
// The revoke is only as strong as the authenticator. In production
// that is auth.HMACAuthenticator, whose Revoke is advisory (ADR 0044
// decision 3): the token stays valid until it expires, and logout
// ends the session by clearing the cookie here and the client's
// stored copy. TestLogoutWithStatelessSessions pins that. For a
// session with a user, POST /logout/everywhere (revocation.go) is the
// stronger version, and does kill every copy.
func logout(c Config, w http.ResponseWriter, r *http.Request) error {
	if cred := auth.CredentialFromRequest(r); cred != "" {
		// Ignore Revoke errors: the stateless authenticator always
		// returns nil, a stateful one may return "unknown token"
		// which we treat as already-revoked. Either way the client
		// just wants its cookie cleared.
		_ = c.Auth.Revoke(r.Context(), cred)
	}
	clearSessionCookie(c, w)
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

// maxJSONBodyBytes caps the request bodies decodeJSON accepts. The
// shapes routed through it (login, create, join, spectate) are a few
// hundred bytes; 64 KiB is generous headroom while denying a hostile
// client an unbounded stream at the decoder. Deck upload has its own
// larger maxDeckBodyBytes limit.
const maxJSONBodyBytes = 64 << 10

// decodeJSON enforces Content-Type: application/json when a body is
// present and rejects unknown fields so typos in request shapes
// surface early. Returns a 400-bearing error (413 when the body
// exceeds maxJSONBodyBytes).
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if r.Body == nil {
		return httpError(http.StatusBadRequest, "missing body")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return httpError(http.StatusRequestEntityTooLarge, fmt.Sprintf("body exceeds %d-byte limit", maxJSONBodyBytes))
		}
		return httpError(http.StatusBadRequest, fmt.Sprintf("invalid body: %s", err.Error()))
	}
	return nil
}

// decodeJSONString decodes a JSON document already held in memory —
// the `report` part of a multipart bug report. Same strictness as
// decodeJSON (unknown fields rejected) minus the Content-Type and
// body-size concerns, which the multipart parser has already handled.
func decodeJSONString(raw string, dst any) error {
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
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

// secureCookies decides the session cookie's Secure attribute.
// CMDCTRL_SECURE_COOKIES, when set non-blank, is an explicit
// truthy/falsy override. Unset, the default follows the deployment's
// own TLS signal: the Discord OAuth redirect URI is configured with
// the public origin, so an https:// prefix there means the server is
// reachable over TLS and the cookie should be Secure. Local http dev
// (no redirect URI, or an http:// one) keeps Secure off so the
// cookie still flows. Read per request, matching the other CMDCTRL
// runtime knobs.
func secureCookies(c Config) bool {
	if v, ok := os.LookupEnv("CMDCTRL_SECURE_COOKIES"); ok && strings.TrimSpace(v) != "" {
		return envflag.Truthy(v)
	}
	return strings.HasPrefix(strings.ToLower(c.Discord.RedirectURI), "https://")
}

// setSessionCookie stamps the browser session cookie that transports
// the credential on subsequent HTTP and same-origin WS requests.
// HttpOnly prevents JS access; SameSite=Lax lets the invite-link
// flow work (top-level navigation). Secure is env-driven via
// secureCookies — prod (HTTPS since S12) gets it by default through
// the https redirect-URI heuristic.
func setSessionCookie(c Config, w http.ResponseWriter, tok string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookie,
		Value:    tok,
		Path:     "/",
		Expires:  exp,
		HttpOnly: true,
		Secure:   secureCookies(c),
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSessionCookie emits a Set-Cookie that evicts the browser's
// current session cookie. MaxAge=-1 tells browsers to drop it
// immediately; Expires in the past covers older clients that ignore
// MaxAge. Secure mirrors setSessionCookie so the eviction targets
// the same cookie variant the login flow set.
func clearSessionCookie(c Config, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secureCookies(c),
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
	case errors.Is(err, ErrInvalidInvite), errors.Is(err, ErrInvalidReclaim):
		status = http.StatusUnauthorized
	case errors.Is(err, ErrGameFull),
		errors.Is(err, ErrGameStarted),
		errors.Is(err, ErrSeatTaken),
		errors.Is(err, ErrAlreadySeated),
		errors.Is(err, ErrDeckNotUploaded),
		errors.Is(err, game.ErrNotEnoughPlayers),
		errors.Is(err, game.ErrGameAlreadyStarted),
		errors.Is(err, game.ErrGameNotInLobby):
		status = http.StatusConflict
	case errors.Is(err, ErrPlayerNotInGame):
		status = http.StatusForbidden
	case errors.Is(err, ErrNotTableManager), errors.Is(err, ErrNotInviteManager):
		status = http.StatusForbidden
	case errors.Is(err, ErrHostIneligible):
		status = http.StatusUnprocessableEntity
	// ADR 0075 §2.3, the two halves of a refused settings patch. A
	// value outside its range is a malformed request (400); a
	// starting-life change after Start is a well-formed request the
	// game's state cannot satisfy (422), which is the documented
	// rejection and the one a client is expected to pre-empt by
	// disabling the control.
	case errors.Is(err, game.ErrInvalidSetting):
		status = http.StatusBadRequest
	case errors.Is(err, game.ErrStartingLifeLocked):
		status = http.StatusUnprocessableEntity
	case errors.Is(err, ErrGameNotActiveForSpawn):
		status = http.StatusConflict
	case errors.Is(err, ErrNotABot), errors.Is(err, ErrUnknownBotTier), errors.Is(err, ErrSeatIsBot):
		status = http.StatusUnprocessableEntity
	case errors.Is(err, ErrGameArchived):
		status = http.StatusConflict
	case errors.Is(err, ErrTooManyReclaims):
		status = http.StatusTooManyRequests
	case errors.Is(err, ErrEmptyName), errors.Is(err, ErrInvalidInviteKind):
		status = http.StatusBadRequest
	case errors.Is(err, auth.ErrInvalidCredential),
		errors.Is(err, auth.ErrExpiredCredential),
		errors.Is(err, auth.ErrRevokedCredential),
		errors.Is(err, auth.ErrUnknownPrincipal):
		status = http.StatusUnauthorized
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// constantTimeEqual compares two strings in constant time via the
// stdlib primitive (a length mismatch still short-circuits inside
// ConstantTimeCompare — that length signal is fine for admin-token
// comparison).
func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// compile-time assertion we haven't dropped context-awareness from
// critical methods.
var _ = context.Background
