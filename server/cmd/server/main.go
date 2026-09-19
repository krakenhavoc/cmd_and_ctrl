// Command server is the cmd_and_ctrl game server. As of S04 it exposes:
//
//	GET  /healthz          — liveness probe
//	POST /admin/login      — exchange admin token for an admin session
//	POST /games            — admin-only: create a new game
//	GET  /games            — authenticated: list games
//	GET  /games/{id}       — authenticated: game metadata
//	POST /games/{id}/join  — public (invite-gated): seat a player + mint session
//	POST /games/{id}/start — authenticated: transition lobby → active
//	GET  /me               — authenticated: echo principal
//	GET  /ws               — WebSocket endpoint speaking the v0 protocol
//
// Configuration is via environment variables:
//
//	CMDCTRL_ADDR          — listen address, default ":8080"
//	CMDCTRL_DATA_DIR      — root for crash-recovery snapshots, default "./data"
//	                        set to empty string to disable disk writes entirely
//	CMDCTRL_ADMIN_TOKEN   — shared secret for the admin login flow. REQUIRED.
//	                        Set to a long random string; anyone with this can
//	                        create games and list the whole lobby.
//	CMDCTRL_SESSION_TTL   — session lifetime, Go duration string (e.g. 12h).
//	                        Default 12h.
//	CMDCTRL_SESSION_KEY   — HMAC-SHA256 key that signs session tokens, at
//	                        least 32 bytes, distinct from the admin token.
//	                        Set: sessions survive a restart. Unset: sessions
//	                        are in memory and die with the process, with a
//	                        warning on every boot. Too short: boot fails.
//	                        Rotating it logs everyone out.
//	CMDCTRL_IDENTITY_KEY  — AES-256-GCM key (same format as the session key:
//	                        a random string of at least 32 bytes, distinct
//	                        from the admin token and the session key) that
//	                        encrypts Discord refresh tokens in the database
//	                        (ADR 0051 decision 5). Unset: sign-in still
//	                        works, the refresh token is discarded and stored
//	                        as NULL, with a warning on every boot. Too short
//	                        or reused: boot fails. Rotating it makes stored
//	                        refresh tokens unreadable; nothing else breaks.
//	CMDCTRL_SEED_DEMO    — if "1", seed a 4-player demo game at startup.
//	                        Useful for the gamecli dev loop when you want a
//	                        ready-to-go room without going through the lobby.
//
// Persistent database (S34 sub-PR 1, ADR 0051). Opened at
// <CMDCTRL_DATA_DIR>/db/cmdctrl.sqlite, before RestoreFromDisk runs;
// skipped entirely when CMDCTRL_DATA_DIR is empty, same as every other
// disk-backed store below.
//
//	CMDCTRL_DB_BACKUP_INTERVAL — Go duration between VACUUM INTO backup
//	                              sweeps (db/cmdctrl.backup.sqlite,
//	                              beside the live file). Default 1h.
//	                              <= 0 disables the sweep. This is the
//	                              in-process, same-disk copy only — the
//	                              nightly off-node copy to HomeLab is
//	                              the deploy owner's job, not this
//	                              server's.
//
// Bot seats (S31). The `random` and `heuristic` tiers need nothing;
// `assisted` and `strong` need a model endpoint, and report themselves
// UNAVAILABLE in the picker until one is configured — a bot labelled
// `assisted` that is quietly playing the heuristic is the failure the
// tier system exists to prevent.
//
//	CMDCTRL_OPENAI_ENDPOINT    — an OpenAI-compatible /v1/chat/completions
//	                             server: Ollama, LM Studio, llama.cpp's
//	                             server, vLLM. A URL (e.g.
//	                             http://192.168.1.18:11434), or "1" for a
//	                             stock Ollama on this machine. Setting it
//	                             enables the model tiers and takes
//	                             precedence over the Anthropic key.
//	CMDCTRL_OPENAI_API_KEY     — optional; most local servers need none.
//	CMDCTRL_OPENAI_SEND_THINK  — "0" stops the client sending Ollama's
//	                             `think: false`. Only for a server that
//	                             rejects the field: leaving a hybrid-
//	                             thinking model (qwen3, …) on its default
//	                             spends the whole deadline thinking and
//	                             answers nothing.
//	CMDCTRL_ANTHROPIC_API_KEY  — hosted alternative (falls back to
//	                             ANTHROPIC_API_KEY). CMDCTRL_ANTHROPIC_ENDPOINT
//	                             overrides the API URL.
//	CMDCTRL_BOT_MODEL          — model id for routine windows. REQUIRED for a
//	                             local endpoint (the served model's name);
//	                             optional for Anthropic, which has defaults.
//	CMDCTRL_BOT_FRONTIER_MODEL — model id for escalated windows. Defaults to
//	                             CMDCTRL_BOT_MODEL: one local model for both
//	                             slots is a supported configuration.
//	CMDCTRL_BOT_MAX_THINK      — Go duration; the model tiers' hard think
//	                             deadline. Default 2s (5s for `strong`) on a
//	                             hosted model, 20s when a local endpoint is
//	                             configured, because a model that overruns
//	                             the deadline plays the heuristic's move.
//	CMDCTRL_BOT_DECISION_LOG   — directory for the per-game bot decision log
//	                             (prompt, reply, ranking, fallback per
//	                             window). Empty (the default) is OFF. The
//	                             file aggregates every bot seat's own view of
//	                             one table, so it is operator-only: never
//	                             served, never attached to a bug report.
//	CMDCTRL_BOT_DECISION_LOG_MODE — escalated (default) | all | model.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/deckprofile"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/bugstore"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"

	// Blank import: effects/wire.go's init() populates the S14
	// EffectResolver / ETBEffectHook / IsCatalogCard callbacks on
	// the game package. Without this import the catalog stays cold
	// and every card falls through to manual sandbox resolution.
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/catalog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/github"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/appenv"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/envflag"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

const (
	// scryfallPath is the relative location of the bulk dump inside
	// CMDCTRL_DATA_DIR. Matches scripts/scryfall-refresh.sh.
	scryfallSubdir      = "scryfall"
	scryfallDefaultFile = "default-cards.json"
	imagesSubdir        = "images"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	// Dev convenience: warn loudly when deck validation is being
	// skipped. Catches the case where the env var leaks into a
	// production deployment — a WARN on every boot is harder to
	// miss than a quiet behaviour change.
	if os.Getenv("CMDCTRL_DEV_SKIP_DECK_VALIDATION") != "" {
		log.Warn("CMDCTRL_DEV_SKIP_DECK_VALIDATION is set; deck validation is BYPASSED. Do not use in production.")
	}
	// Same family: the e2e suite relaxes the lobby rate limiters via
	// this knob (see lobby.Handler). Loud on boot for the same
	// leaked-into-prod reason.
	if envflag.Truthy(os.Getenv("CMDCTRL_DEV_RELAX_RATE_LIMITS")) {
		log.Warn("CMDCTRL_DEV_RELAX_RATE_LIMITS is set; lobby rate limits are EFFECTIVELY DISABLED. Do not use in production.")
	}

	cfg := loadConfig(log)

	// Say which deployment this is on every boot. The dev banner in
	// the client comes from the same value, so a mismatch between the
	// journal and the browser is the first thing to check when
	// "why is this feature missing on dev" comes up.
	if cfg.Env.IsDev() {
		log.Warn("running in DEVELOP environment; dev-only features are exposed. Never point this at production data.",
			"env", cfg.Env.String(), "features", fmt.Sprintf("%+v", cfg.Features))
	} else {
		log.Info("running in production environment", "env", cfg.Env.String())
	}
	// A CMDCTRL_DEV_* line copied into the prod env file does nothing
	// (appenv refuses to enable features outside dev), but it means
	// someone believes it does — surface it rather than let the
	// misconception sit in the file.
	for _, k := range appenv.StrayProdOverrides(cfg.Env) {
		log.Warn("dev feature override set in a production deployment; it has no effect and should be removed", "var", k)
	}

	// Process-lifetime context: canceled on SIGINT/SIGTERM. Created
	// here, rather than just before srv.Shutdown as before S34, so the
	// database's backup loop (below) can share it and stop on the same
	// signal without its own plumbing.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// The persistent database (ADR 0051, S34 sub-PR 1): people, their
	// games and their decks in later sub-PRs; today just the migrated,
	// WAL-mode file and its backup sweep. Opened and migrated BEFORE
	// RestoreFromDisk below, so a migration failure or a too-new schema
	// (ErrSchemaTooNew) stops the boot before anything reads game state
	// — mirroring game.SnapshotSchemaVersion's refuse-loudly posture.
	//
	// Skipped entirely when CMDCTRL_DATA_DIR is empty, same as the card
	// index, avatar cache and bug-report store below: an empty data dir
	// means "no disk persistence at all" for this deployment.
	var database *db.DB
	if cfg.DataDir != "" {
		var err error
		database, err = db.Open(ctx, cfg.DataDir)
		if err != nil {
			log.Error("database open/migrate failed", "err", err)
			os.Exit(1)
		}
		log.Info("database opened", "path", database.Path())
		go database.RunBackupLoop(ctx, log, cfg.DBBackupInterval)
		if cfg.DBBackupInterval > 0 {
			log.Info("database backup sweep started", "interval", cfg.DBBackupInterval)
		} else {
			log.Warn("CMDCTRL_DB_BACKUP_INTERVAL <= 0; the in-process backup sweep is disabled")
		}
	} else {
		log.Info("CMDCTRL_DATA_DIR is empty; persistent database disabled")
	}

	// Auth + room manager are global singletons for the lifetime of
	// the process. They outlive individual games.
	//
	// Sessions are HMAC-signed when CMDCTRL_SESSION_KEY is set, so a
	// token survives a deploy (ADR 0044 decision 3, #517). Unset falls
	// back to the in-memory store with a loud warning; a key that is
	// set but too short fails the boot. There is no default key.
	authenticator := newAuthenticator(log, cfg)
	mgr := ws.NewRoomManager(log, cfg.DataDir)
	// With a database, games / seats / invites are rows (ADR 0051
	// decision 4, S34 sub-PR 3) and RestoreFromDisk imports any
	// lobby/*.json the previous binary left. Without one, the lobby
	// keeps its metadata in memory and nothing about a game survives
	// the process — the same as every other artifact under an empty
	// CMDCTRL_DATA_DIR.
	var l *lobby.Lobby
	if database != nil {
		l = lobby.NewLobbyWithStore(mgr, lobby.NewSQLStore(database))
	} else {
		l = lobby.NewLobby(mgr)
	}

	// People (ADR 0051 decisions 2 and 5, S34 sub-PR 2). A Discord
	// sign-in upserts a users + identities row and the session carries
	// the user's id. Without a database there is nowhere to put them,
	// and sign-in mints a session with a zero UserID, as before.
	userStore := newUserStore(log, database)

	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&lobby.WSAuthorizer{Auth: authenticator})
	// Lobby HTTP mutations (join/deck/start) broadcast through the
	// hub so clients already on the game page see them immediately.
	l.SetStateBroadcaster(hub)
	// S31: bot seats. The manager starts a runner per bot seat when a
	// game starts (and when one is restored below) and stops them
	// when the game is deleted or the process exits; runners
	// broadcast their moves through the hub like any other commit.
	// Wired BEFORE RestoreFromDisk so a resumed game with a bot seat
	// gets its runner back rather than hanging on an empty chair.
	//
	// The POLICY factory is injected further down, after the card
	// index has loaded — the model tiers build their prompts from it.
	// That injection also happens before RestoreFromDisk, so a
	// resumed `heuristic` seat comes back as a heuristic seat.
	bots := aiseat.NewManager(hub, log)
	l.SetBotHost(bots)
	if len(cfg.AllowedOrigins) > 0 {
		hub.SetAllowedOrigins(cfg.AllowedOrigins)
		log.Info("ws allowed-origins configured", "hosts", cfg.AllowedOrigins)
	}

	// Card index + image cache. The Scryfall bulk dump is optional
	// at startup: a missing file logs a warning and the /cards*
	// routes serve 404 / 503 until the next scryfall-refresh run.
	// This keeps the server bootable on a fresh deployment before
	// the cron has landed its first dump.
	cardIdx, imgCache := loadCardAssets(log, cfg.DataDir)

	// The bot policy factory. This is the line that decides which
	// tiers this deployment can seat: aiseat itself can only build
	// `random` (every policy package imports it, so it cannot import
	// them back — see aiseat.PolicyFactory), and aiseat/tiers, which
	// sits above all of them, can build all four. main is the only
	// place above both, which is why the wiring is here.
	//
	// It runs after the card index because a model tier's prompt
	// carries its own decklist with oracle text, and before
	// RestoreFromDisk because a restored bot seat relaunches its
	// runner and must get the policy its tier names.
	bots.SetPolicyFactory(botFactory(log, cfg, cardIdx))
	// The decision log, when an operator asked for one. Off by
	// default; a failure to open the directory is logged and the
	// server boots without it, because a diagnostic is never worth a
	// refused boot.
	if cfg.BotDecisionLog != "" {
		dl, err := decisionlog.New(decisionlog.Options{
			Dir:  cfg.BotDecisionLog,
			Mode: cfg.BotDecisionLogMode,
			Log:  log,
		})
		if err != nil {
			log.Error("bot decision log could not be started; bots play without one", "err", err)
		} else {
			bots.SetDecisionLogger(dl.DecisionLogger())
			log.Warn("BOT DECISION LOG IS ON: every bot seat's own view of every game is written to disk. Operator-only — it is never served and must never be attached to a bug report.",
				"dir", dl.Dir(), "mode", string(dl.Mode()))
		}
	}

	// Restore games that were live when the previous process exited.
	// This is the read half of persistence — see internal/game/
	// snapshot.go and internal/ws/persist.go.
	//
	// It runs here, before any route is registered, so no client can
	// observe a half-built lobby. It runs after the card assets load
	// because a restored board is projected through the same view
	// path a live one is; the card-effect catalog it also needs is
	// wired by the effects package's init, which has already fired.
	//
	// A game this binary cannot rebuild is abandoned rather than
	// guessed at — the players get the "game not found" they would
	// have got from any restart before this feature existed. Nothing
	// here is fatal: a bad restore point must never stop a boot.
	//
	// NOTE: auth sessions survive a restart only when
	// CMDCTRL_SESSION_KEY is set (auth.HMACAuthenticator). Without it
	// they are in memory and die with the process, and players
	// re-authenticate through their invite link. That link is why
	// lobby metadata is persisted alongside the engine snapshot.
	//
	// The Scryfall lookup is wired first: a restore point written
	// before #683 has no Card.VariableToughness, and restore
	// recomputes it from the printing (game/snapshot_backfill.go).
	// With no index it falls back to the pre-#683 rule.
	game.PrintedVariableToughness = deck.PrintedVariableToughness(cardIdx)
	if n := l.RestoreFromDisk(log); n > 0 {
		log.Info("resumed games from the previous process", "count", n)
	}

	// Optional demo game for the gamecli dev path. Creates a game
	// directly (bypassing the lobby's invite flow) so you can dial
	// the ws endpoint with any admin token and a ?game=<demo-id>
	// query param.
	if cfg.SeedDemo {
		g := seedDemoGame(log)
		mgr.Create(g)
		log.Info("seeded demo game", "id", g.ID.String())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /ws", hub.ServeWS)
	mux.Handle("/cards/", auth.Middleware(authenticator)(cards.Handler(cardIdx, imgCache)))
	// The card catalog: a browsable list of every card the engine
	// actually automates, with each entry's declared completeness.
	//
	// Behind auth.Middleware, like /cards/. It was built unauthenticated
	// as a public showcase, and it is gated because AGENTS.md §1 and §8
	// describe this project as private and personal-use: serving card
	// art to anonymous visitors is a different posture from the one the
	// repo states, and the page is no less useful to a signed-in
	// player. Not behind requireDev — it ships in production, it is
	// simply not public.
	//
	// catalog.Handler's own doc explains why its image route is scoped
	// to registered cards rather than proxying all ~35k Scryfall UUIDs;
	// that scoping still matters, since a session is cheap to obtain.
	// Both patterns are more specific than "/", so the lobby catch-all
	// below does not shadow them.
	mux.Handle("GET /catalog", auth.Middleware(authenticator)(catalog.Handler(cardIdx, imgCache)))
	mux.Handle("/catalog/", auth.Middleware(authenticator)(catalog.Handler(cardIdx, imgCache)))
	discordCfg := discord.ConfigFromEnv()
	if discordCfg.Enabled() {
		log.Info("discord oauth enabled", "redirect_uri", discordCfg.RedirectURI)
	} else {
		log.Info("discord oauth disabled — set CMDCTRL_DISCORD_CLIENT_ID/SECRET/REDIRECT_URI to enable")
	}

	// Avatar cache lives under the same data dir as the scryfall
	// index and replay dumps. Empty DataDir disables the cache —
	// the /avatars endpoint will then 503 and the client falls
	// back to its initials placeholder.
	var avatarDir string
	if cfg.DataDir != "" {
		avatarDir = filepath.Join(cfg.DataDir, "avatars")
	}
	avatarCache := discord.NewAvatarCache(avatarDir, nil)

	// Bug reporting (in-app "report a bug" button → GitHub issue).
	// Token unset = feature off; the client hides the button via
	// GET /bugreport/config. The token should be a fine-grained PAT
	// with Issues:write on the one repo — see ADR 0017.
	var bugReporter lobby.BugReporter
	ghRepo := os.Getenv("CMDCTRL_GITHUB_REPO")
	if ghRepo == "" {
		ghRepo = "krakenhavoc/cmd_and_ctrl"
	}
	if ghToken := os.Getenv("CMDCTRL_GITHUB_TOKEN"); ghToken != "" {
		bugReporter = github.NewClient(ghToken, ghRepo)
		log.Info("bug reporting enabled", "repo", ghRepo)
	} else {
		log.Info("bug reporting disabled — set CMDCTRL_GITHUB_TOKEN to enable")
	}

	// Bug-report artifacts: screenshots and pinned replays, under the
	// same data dir as the scryfall index, avatars, and replay dumps.
	//
	// Attachments additionally need a PUBLIC base URL, because a
	// GitHub issue can only render an image it can fetch — see ADR
	// 0017 §6. There is deliberately no default: a wrong origin
	// produces issues full of broken images, which is worse than a
	// deploy where the modal simply doesn't offer file upload.
	var bugDir string
	if cfg.DataDir != "" {
		bugDir = filepath.Join(cfg.DataDir, "bugreports")
	}
	publicBase := os.Getenv("CMDCTRL_PUBLIC_BASE_URL")
	if publicBase == "" {
		publicBase = os.Getenv("CMDCTRL_CLIENT_BASE_URL")
	}
	bugStore := bugstore.New(bugDir, publicBase)
	if bugStore.Enabled() {
		if bugStore.AttachmentsEnabled() {
			log.Info("bug report attachments enabled", "dir", bugDir, "public_base_url", publicBase)
		} else {
			log.Info("bug report attachment hosting disabled — set CMDCTRL_PUBLIC_BASE_URL to enable screenshots; replay pinning remains enabled", "dir", bugDir)
		}
		// Enforce retention once at boot as well as after each report,
		// so a server that files nothing for months still reclaims the
		// artifacts of the reports it filed before.
		if n, err := bugStore.Prune(bugstore.DefaultRetention, bugstore.DefaultMaxStoreBytes); err != nil {
			log.Warn("pruning old bug reports failed", "err", err)
		} else if n > 0 {
			log.Info("pruned expired bug reports", "count", n)
		}
	} else if bugReporter != nil {
		log.Info("bug report attachments disabled — needs CMDCTRL_DATA_DIR and CMDCTRL_PUBLIC_BASE_URL; text reports still work")
	}

	mux.Handle("/", lobby.Handler(lobby.Config{
		Lobby:             l,
		Auth:              authenticator,
		AdminToken:        cfg.AdminToken,
		SessionTTL:        cfg.SessionTTL,
		Env:               cfg.Env,
		Features:          cfg.Features,
		Cards:             cardIdx,
		Evictor:           hub,
		Discord:           discordCfg,
		DiscordStateStore: discord.NewStateStore(),
		DiscordAvatars:    avatarCache,
		Users:             userStore,
		BugReporter:       bugReporter,
		BugStore:          bugStore,
		Log:               log,
		Bots:              bots,
		// The four curated archetype decks from S31 sub-PR 5. This
		// replaced aiseat.PlaceholderDecks() — the ninety-nine
		// Mountains stand-in sub-PR 4 shipped while the real decks
		// were still being built. See botdecks.go.
		BotDecks: botDeckCatalog{},
	}))

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		// Reap idle keep-alive connections so abandoned sockets don't
		// pin fds indefinitely. NO WriteTimeout on purpose: it would
		// sever long-lived WebSockets and replay streams mid-flight.
		IdleTimeout: 120 * time.Second,
	}

	go func() {
		log.Info("server listening", "addr", cfg.Addr, "data_dir", cfg.DataDir)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Order matters: close the HTTP listener first so no new WebSocket
	// upgrades can slip in after we've taken a snapshot of live clients.
	// ServeWS is a hijacking handler that returns immediately after
	// spawning its pumps, so srv.Shutdown itself completes quickly —
	// it's the act of closing the listener that we need.
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown", "err", err)
	}
	hub.Shutdown(shutdownCtx)
	bots.Shutdown()
	if database != nil {
		if err := database.Close(); err != nil {
			log.Error("database close", "err", err)
		}
	}
	log.Info("server stopped")
}

// config is the minimal set of env-derived values main() needs. Kept
// in its own struct so that validation (in loadConfig) is a single
// call site rather than scattered os.Getenv checks.
type config struct {
	Addr       string
	DataDir    string
	AdminToken string
	SessionTTL time.Duration
	SeedDemo   bool
	// AllowedOrigins is the cross-origin hostname allow-list passed to
	// the hub's WebSocket CheckOrigin. Same-origin is always allowed
	// (no config needed). Set CMDCTRL_ALLOWED_ORIGINS to a comma-
	// separated list for LAN clients reaching the server from a
	// different host/port than the one it binds on.
	AllowedOrigins []string
	// BotMaxThink overrides the hard think deadline the model bot
	// tiers get (CMDCTRL_BOT_MAX_THINK). Zero takes ADR 0033 §10's
	// defaults for a hosted model, or a larger one for a self-hosted
	// endpoint — see botFactory.
	BotMaxThink time.Duration
	// BotModel and BotFrontierModel are the model ids the funnel
	// asks for (CMDCTRL_BOT_MODEL / CMDCTRL_BOT_FRONTIER_MODEL).
	// Empty keeps the shipped Anthropic defaults; a local endpoint
	// needs at least BotModel, and the two may name the same model.
	BotModel         string
	BotFrontierModel string
	// BotDecisionLog is the directory the per-game bot decision log
	// is written to (CMDCTRL_BOT_DECISION_LOG). Empty is off, which
	// is the default: the file holds every bot seat's view of one
	// table and is operator-only. BotDecisionLogMode is its fullness.
	BotDecisionLog     string
	BotDecisionLogMode decisionlog.Mode
	// DBBackupInterval is how often the persistent database's VACUUM
	// INTO backup sweep runs (CMDCTRL_DB_BACKUP_INTERVAL). Defaults to
	// db.DefaultBackupInterval; <= 0 disables the sweep.
	DBBackupInterval time.Duration
	// Env is the deployment identity from CMDCTRL_ENV. Unset means
	// production — a forgotten variable fails closed.
	Env appenv.Env
	// Features are the dev-only capabilities this deployment exposes,
	// derived entirely from Env. Always zero in production.
	Features appenv.Features
}

// loadConfig pulls the server's env vars, applies defaults, and
// exits (with a user-facing log message) if a required value is
// missing. This runs BEFORE the server does any work so misconfig
// shows up immediately instead of 500-ing the first login attempt.
func loadConfig(log *slog.Logger) config {
	env, err := appenv.Parse(os.Getenv(appenv.EnvVar))
	if err != nil {
		// Fatal rather than defaulting: silently treating an
		// unrecognised value as prod would strip every dev feature off
		// the dev box with no signal but a missing button.
		log.Error("invalid deployment environment", "err", err)
		os.Exit(1)
	}

	c := config{
		Addr:             envOr("CMDCTRL_ADDR", ":8080"),
		AdminToken:       os.Getenv("CMDCTRL_ADMIN_TOKEN"),
		SeedDemo:         os.Getenv("CMDCTRL_SEED_DEMO") == "1",
		SessionTTL:       12 * time.Hour,
		BotDecisionLog:   strings.TrimSpace(os.Getenv("CMDCTRL_BOT_DECISION_LOG")),
		BotModel:         strings.TrimSpace(os.Getenv("CMDCTRL_BOT_MODEL")),
		BotFrontierModel: strings.TrimSpace(os.Getenv("CMDCTRL_BOT_FRONTIER_MODEL")),
		DBBackupInterval: db.DefaultBackupInterval,
		Env:              env,
		Features:         appenv.LoadFeatures(env),
	}

	// CMDCTRL_DB_BACKUP_INTERVAL is a misconfiguration worth failing
	// the boot over, same reasoning as CMDCTRL_BOT_MAX_THINK below: a
	// deployment that asked for a backup cadence and silently did not
	// get one is worse than a boot that refuses to start.
	if raw := os.Getenv("CMDCTRL_DB_BACKUP_INTERVAL"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			log.Error("CMDCTRL_DB_BACKUP_INTERVAL invalid", "value", raw, "err", err)
			os.Exit(1)
		}
		c.DBBackupInterval = d
	}

	// CMDCTRL_BOT_MAX_THINK is a misconfiguration worth failing the
	// boot over rather than ignoring: a deployment that asked for a
	// longer bot deadline and silently did not get one produces
	// heuristic play labelled `assisted`, which is the one outcome
	// the bot tiers are built to prevent.
	if raw := os.Getenv("CMDCTRL_BOT_MAX_THINK"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			log.Error("CMDCTRL_BOT_MAX_THINK invalid", "value", raw, "err", err)
			os.Exit(1)
		}
		if d <= 0 {
			log.Error("CMDCTRL_BOT_MAX_THINK must be positive", "value", raw)
			os.Exit(1)
		}
		c.BotMaxThink = d
	}

	// The mode only matters when the log is ON, and it fails the boot
	// only then. Same posture as CMDCTRL_BOT_MAX_THINK for a
	// deployment that asked for a log — silently getting a mode it
	// did not ask for means discovering it after a night of games —
	// but a stale CMDCTRL_BOT_DECISION_LOG_MODE left in an env file
	// beside an unset log is refusing to start over a variable that
	// changes nothing. That is a warning, not a dead server.
	rawMode := os.Getenv("CMDCTRL_BOT_DECISION_LOG_MODE")
	mode, merr := decisionlog.ParseMode(rawMode)
	switch {
	case merr == nil:
		c.BotDecisionLogMode = mode
	case c.BotDecisionLog != "":
		log.Error("CMDCTRL_BOT_DECISION_LOG_MODE invalid", "value", rawMode, "err", merr)
		os.Exit(1)
	default:
		log.Warn("CMDCTRL_BOT_DECISION_LOG_MODE is not a mode this build knows, and is being ignored because CMDCTRL_BOT_DECISION_LOG is unset (the bot decision log is off)",
			"value", rawMode, "err", merr)
		c.BotDecisionLogMode = decisionlog.ModeEscalated
	}

	if raw := os.Getenv("CMDCTRL_SESSION_TTL"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			log.Error("CMDCTRL_SESSION_TTL invalid", "value", raw, "err", err)
			os.Exit(1)
		}
		if d <= 0 {
			log.Error("CMDCTRL_SESSION_TTL must be positive", "value", raw)
			os.Exit(1)
		}
		c.SessionTTL = d
	}

	// CMDCTRL_DATA_DIR: honour an explicit empty string (disables
	// disk writes). Unset → "./data".
	if _, present := os.LookupEnv("CMDCTRL_DATA_DIR"); present {
		c.DataDir = os.Getenv("CMDCTRL_DATA_DIR")
	} else {
		c.DataDir = "./data"
	}

	if raw := os.Getenv("CMDCTRL_ALLOWED_ORIGINS"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			if s = strings.TrimSpace(s); s != "" {
				c.AllowedOrigins = append(c.AllowedOrigins, s)
			}
		}
	}

	if c.AdminToken == "" {
		log.Error("CMDCTRL_ADMIN_TOKEN is required — set it to a long random string before starting the server")
		os.Exit(1)
	}
	if len(c.AdminToken) < 16 {
		log.Error("CMDCTRL_ADMIN_TOKEN is too short — use at least 16 characters")
		os.Exit(1)
	}
	return c
}

// newAuthenticator picks the session store (see auth.NewFromEnv) and
// exits on a misconfigured key rather than booting on a weaker one.
//
// The session key must not be the admin token. The admin token is
// copied into the Discord bot's env file on every deploy (ADR 0004
// §6), so a shared value would let anything that can read that file
// forge a session for any seat.
func newAuthenticator(log *slog.Logger, cfg config) auth.Authenticator {
	if strings.TrimSpace(os.Getenv(auth.SessionKeyEnv)) == cfg.AdminToken {
		log.Error(auth.SessionKeyEnv + " must not equal CMDCTRL_ADMIN_TOKEN; generate a separate random key")
		os.Exit(1)
	}
	a, err := auth.NewFromEnv(os.Getenv, log)
	if err != nil {
		log.Error("session key invalid", "var", auth.SessionKeyEnv, "err", err)
		os.Exit(1)
	}
	if _, ok := a.(*auth.HMACAuthenticator); ok {
		log.Info("sessions are HMAC-signed and survive a restart", "var", auth.SessionKeyEnv)
	}
	return a
}

// newUserStore builds the user store and its refresh-token key, and
// exits on a misconfigured key rather than booting on a weak or
// shared one (users.NewSealerFromEnv). The key is checked whether or
// not there is a database, as the session key is: a bad value in the
// env file is an operator mistake worth hearing about on any boot.
//
// With the key absent the store still works and refresh tokens are
// discarded (ADR 0051 decision 5); NewSealerFromEnv has already
// warned, naming the variable.
func newUserStore(log *slog.Logger, database *db.DB) users.Store {
	sealer, err := users.NewSealerFromEnv(os.Getenv, log)
	if err != nil {
		log.Error("identity key invalid", "var", users.IdentityKeyEnv, "err", err)
		os.Exit(1)
	}
	if database == nil {
		return users.NoStore{}
	}
	if sealer != nil {
		log.Info("Discord refresh tokens are stored encrypted", "var", users.IdentityKeyEnv)
	}
	return users.NewSQLStore(database, sealer)
}

func envOr(key, dflt string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return dflt
}

// botModelDefaults are the pacing decisions that depend on WHICH
// model transport a deployment configured.
const (
	// localDefaultMaxThink is the think deadline a self-hosted model
	// gets when the deployment did not name one. ADR 0033 §10's 2s
	// was sized for a hosted cheap model; a 7B model on a desktop GPU
	// misses it on most windows, and a model tier that misses its
	// deadline every window is Layer B wearing a stronger name. 20s
	// is slow for a four-player table and honest about what it is.
	localDefaultMaxThink = 20 * time.Second
	// localMinSaneMaxThink is the line under which a self-hosted
	// deployment gets a warning: below this, expect the heuristic.
	localMinSaneMaxThink = 5 * time.Second
)

// botFactory builds the policy factory for the bot seats, and logs
// what this deployment can therefore offer.
//
// Transport selection, in order:
//
//   - CMDCTRL_OPENAI_ENDPOINT — an OpenAI-compatible /v1/chat/
//     completions server. This is the local-LLM path (Ollama, LM
//     Studio, llama.cpp's server, vLLM) and it wins when set,
//     because naming a specific endpoint is a more deliberate act
//     than leaving an API key in the environment.
//   - CMDCTRL_ANTHROPIC_API_KEY (or ANTHROPIC_API_KEY) — the hosted
//     path.
//   - neither — `random` and `heuristic` only. Not a broken server:
//     a smaller one, which says so in the picker.
func botFactory(log *slog.Logger, cfg config, idx *cards.Index) *tiers.Factory {
	var client model.Client
	local := false
	switch oc, ac := model.NewOpenAIClient(), model.NewAnthropicClient(); {
	case oc != nil:
		client, local = oc, true
		log.Info("bot model transport: OpenAI-compatible (local LLM)",
			"endpoint", oc.URL(), "authenticated", oc.APIKey != "", "thinking_suppressed", !oc.OmitThink)
	case ac != nil:
		client = ac
		log.Info("bot model transport: Anthropic")
	default:
		log.Info("bot model tiers disabled — set CMDCTRL_OPENAI_ENDPOINT (a local LLM) or CMDCTRL_ANTHROPIC_API_KEY to enable `assisted` and `strong`")
	}

	maxThink := cfg.BotMaxThink
	if local && maxThink == 0 {
		maxThink = localDefaultMaxThink
		log.Info("bot think deadline raised for the local model transport; set CMDCTRL_BOT_MAX_THINK to choose your own",
			"max_think", maxThink)
	}
	// The silent-downgrade warning. A local model that cannot answer
	// inside the deadline falls back to the heuristic on EVERY
	// window, which looks from the table exactly like an `assisted`
	// bot playing badly. Say it once at boot rather than leave it to
	// be inferred from a per-window WARN.
	if local && maxThink > 0 && maxThink < localMinSaneMaxThink {
		log.Warn("CMDCTRL_BOT_MAX_THINK is short for a self-hosted model; windows that overrun it play the HEURISTIC's move under the model tier's name. Watch for 'bot model call TIMED OUT' lines.",
			"max_think", maxThink)
	}

	models := tiers.Models{Routine: cfg.BotModel, Frontier: cfg.BotFrontierModel}
	if local && models.Routine == "" {
		log.Warn("no bot model id configured; set CMDCTRL_BOT_MODEL to the id your endpoint serves (e.g. the name you `ollama pull`ed). Until then the Anthropic defaults are sent and the endpoint will answer 404.")
	}
	if models.SingleModel() {
		// Not a misconfiguration — see tiers.Models. Logged so that
		// an operator reading the metrics knows why the escalation
		// rate no longer changes which model answered.
		log.Info("bot funnel is single-model: routine and frontier windows go to the same model; escalation still buys the wider candidate list",
			"model", models.Routine)
	}

	f := tiers.NewFactory(tiers.FactoryOptions{
		Client:   client,
		Models:   models,
		MaxThink: maxThink,
		DeckProfile: func(deckID string) (model.DeckProfile, bool) {
			return deckprofile.Build(idx, deckID)
		},
	})

	var available []string
	for _, t := range aiseat.Tiers() {
		if st := f.TierStatus(t.Tier); st.Available {
			available = append(available, string(t.Tier))
		}
	}
	log.Info("bot tiers available", "tiers", available, "decks", len(decks.IDs()))
	return f
}

// loadCardAssets boots the card index and image cache from the data
// dir. A missing bulk dump is logged as a warning, not a fatal: the
// server stays usable for lobby-only operations and starts serving
// card images as soon as the next scryfall-refresh writes the file.
//
// Empty DataDir (disk persistence fully disabled) returns a nil
// index and cache; the /cards routes handle that with 404 / 503.
func loadCardAssets(log *slog.Logger, dataDir string) (*cards.Index, *cards.ImageCache) {
	if dataDir == "" {
		log.Warn("CMDCTRL_DATA_DIR is empty; card index and image cache disabled")
		return nil, nil
	}
	idx := cards.NewIndex()
	dumpPath := filepath.Join(dataDir, scryfallSubdir, scryfallDefaultFile)
	if n, err := idx.Load(dumpPath); err != nil {
		if os.IsNotExist(err) {
			log.Warn("scryfall dump not found; run scripts/scryfall-refresh.sh", "path", dumpPath)
		} else {
			log.Error("scryfall dump load failed", "path", dumpPath, "err", err)
		}
	} else {
		log.Info("scryfall index loaded", "cards", n, "path", dumpPath)
	}

	cache, err := cards.NewImageCache(filepath.Join(dataDir, imagesSubdir))
	if err != nil {
		log.Error("image cache init failed", "err", err)
		return idx, nil
	}
	return idx, cache
}

// seedDemoGame builds a 4-player Commander game in the active state,
// each seat with a 99-card filler library plus a placeholder
// commander. This exists so the gamecli dev path has a target
// without going through the lobby flow — turned on via
// CMDCTRL_SEED_DEMO=1.
func seedDemoGame(log *slog.Logger) *game.Game {
	g := game.NewGame()
	names := []string{"Alice", "Bob", "Carol", "Dave"}
	for i, name := range names {
		deck := make([]game.Card, 0, 100)
		deck = append(deck, game.NewCommander(fmt.Sprintf("Demo Commander %d", i+1), uuid.Nil))
		for j := range 99 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %03d", j+1), uuid.Nil))
		}
		p, err := g.AddPlayer(name, deck)
		if err != nil {
			log.Error("seedDemoGame AddPlayer failed", "name", name, "err", err)
			os.Exit(1)
		}
		log.Info("demo player seated", "name", name, "seat", p.Seat, "id", p.ID.String())
	}
	if err := g.Start(nil); err != nil {
		log.Error("seedDemoGame Start failed", "err", err)
		os.Exit(1)
	}
	return g
}
