package aiseat

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// SeatSpec is one bot seat to run: which seat, at what tier, with
// which deck. The lobby builds these from its own seat list and hands
// them to StartBots; keeping the type here rather than in lobby is
// what lets aiseat stay the lower layer, imported by the lobby and
// importing nothing of it.
type SeatSpec struct {
	PlayerID uuid.UUID
	Tier     string
	// Deck is the curated-deck ID the seat was added with, empty when
	// the caller pasted a raw decklist instead. It is carried for the
	// model tiers: the static, prompt-cached half of their prompt is
	// this deck's list and plan, and a PolicyFactory resolves the ID
	// to that profile. A seat with no deck ID still plays — the model
	// simply sees the board and the move list without its own
	// decklist in front of it.
	Deck string
}

// Manager owns the runners for every bot seat on the server: one
// StartBots per game (from Lobby.Start, and from the lobby's restore
// path after a deploy) and one StopBots per game (from Lobby.Delete
// and from shutdown). Satisfies lobby.BotHost. Added in S31 sub-PR 4.
//
// A runner does not need stopping when the game ENDS — it watches the
// room and exits on its own the moment the state leaves StateActive,
// which the final commit wakes it for. StopBots exists for the two
// cases the game never reaches an ending: the game is deleted, and
// the process is going away.
type Manager struct {
	bc  Broadcaster
	log *slog.Logger
	// cfg overrides the per-tier pacing when non-zero. Production
	// leaves it zero and takes ConfigFor(tier); tests set MinThink 0
	// so a whole game runs in milliseconds.
	cfg *Config
	// factory builds the seat policies and is the authority on which
	// tiers this server can play. Nil means builtinFactory: `random`
	// and nothing else. main.go injects aiseat/tiers' factory, which
	// aiseat cannot construct itself without an import cycle — see
	// PolicyFactory.
	factory PolicyFactory

	// decisionLog, when set, opens one decision log per game and
	// installs it as every bot seat's observer. Nil is off, which is
	// the default and what production runs unless an operator asked
	// for it — see aiseat/decisionlog for why.
	decisionLog DecisionLogger

	mu    sync.Mutex
	games map[uuid.UUID]*botGame
}

type botGame struct {
	cancel  context.CancelFunc
	runners []*Runner
	// dlog is this game's decision log, nil when logging is off. It
	// is closed exactly once: by StopBots, or by the watcher
	// goroutine when every runner has exited on its own (a game that
	// simply ENDS is never stopped — the runners see the state leave
	// StateActive and return), or when a second StartBots replaces
	// this set.
	dlog     GameDecisionLog
	closeLog sync.Once
}

// finishLog closes the game's decision log, once.
func (b *botGame) finishLog(log *slog.Logger) {
	if b.dlog == nil {
		return
	}
	b.closeLog.Do(func() {
		if err := b.dlog.Close(); err != nil {
			log.Error("closing the bot decision log failed", "err", err)
		}
	})
}

// NewManager builds a Manager that paces each runner by its tier —
// ConfigFor until a policy factory is injected, and the factory's
// RunnerConfig after (a deployment may have raised the think deadline
// for a slow self-hosted model). bc is the hub and may be nil in
// tests.
func NewManager(bc Broadcaster, log *slog.Logger) *Manager {
	if log == nil {
		log = slog.Default()
	}
	return &Manager{bc: bc, log: log, games: make(map[uuid.UUID]*botGame)}
}

// NewManagerWithConfig is NewManager with one pacing for every tier.
// Tests use it to strip MinThink; production should not.
func NewManagerWithConfig(bc Broadcaster, cfg Config, log *slog.Logger) *Manager {
	m := NewManager(bc, log)
	m.cfg = &cfg
	return m
}

// SetPolicyFactory injects the factory that builds seat policies, and
// with it the answer to "which tiers does this server offer". Call it
// once at boot, BEFORE any game starts or is restored: a runner that
// is already running keeps the policy it was built with.
//
// This is a setter rather than a constructor argument because the
// Manager has to exist before the factory does — main.go wires the
// Manager into the lobby before the card index (which the model
// tiers' deck profiles are built from) has finished loading.
func (m *Manager) SetPolicyFactory(f PolicyFactory) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.factory = f
}

// SetDecisionLogger turns the per-game decision log on. Call it once
// at boot, beside SetPolicyFactory: a game already running keeps the
// observer (or the absence of one) its runners were started with.
//
// Nil turns it off. The log is OFF by default and that is deliberate
// — the file aggregates every bot seat's view of one table, so it is
// operator-only. See aiseat/decisionlog.
func (m *Manager) SetDecisionLogger(dl DecisionLogger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.decisionLog = dl
}

// policyFactory is the injected factory, or the builtin one.
func (m *Manager) policyFactory() PolicyFactory {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.factory == nil {
		return builtinFactory{}
	}
	return m.factory
}

// TierInfo is the picker's view of this server: every declared tier,
// each one marked available or not by the factory that would have to
// build it, with a reason when it cannot.
func (m *Manager) TierInfo() []TierInfo {
	return tiersFor(m.policyFactory())
}

// Tiers lists the policy tiers a bot may be added with. Only the
// available ones — the lobby validates an add request against this,
// and an unavailable tier must be refused rather than downgraded.
func (m *Manager) Tiers() []string {
	var out []string
	for _, t := range m.TierInfo() {
		if t.Available {
			out = append(out, string(t.Tier))
		}
	}
	return out
}

// TierReason explains why a declared tier is not on offer here, for
// the picker to show next to the greyed-out row. Empty for a tier
// that IS available, and for a string that is not a tier at all.
func (m *Manager) TierReason(tier string) string {
	for _, t := range m.TierInfo() {
		if string(t.Tier) == tier {
			return t.Reason
		}
	}
	return ""
}

// configFor is the runner pacing for a tier. The explicit override
// (NewManagerWithConfig, which tests use to strip MinThink) wins;
// otherwise the factory decides, because it is the half that knows
// whether this deployment raised the think deadline for a slow
// self-hosted model. f may be nil.
func (m *Manager) configFor(f PolicyFactory, t Tier) Config {
	if m.cfg != nil {
		return *m.cfg
	}
	if f == nil {
		return ConfigFor(t)
	}
	return f.RunnerConfig(t)
}

// StartBots launches a runner per seat on the room. Calling it twice
// for the same game replaces the first set (the old runners are
// cancelled first), so a restore racing a start cannot double-drive a
// seat. A seat whose tier has no policy is logged and skipped rather
// than quietly played at random — the tier was validated when the bot
// was seated, so reaching here means the build changed under a
// persisted game.
func (m *Manager) StartBots(room *ws.Room, seats []SeatSpec) {
	if room == nil || len(seats) == 0 {
		return
	}
	gameID := room.Game.ID
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.games[gameID]; ok {
		old.cancel()
		for _, r := range old.runners {
			<-r.Done()
		}
		old.finishLog(m.log)
	}
	factory := m.factory // holding m.mu; see policyFactory
	if factory == nil {
		factory = builtinFactory{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	bg := &botGame{cancel: cancel}
	// One log per GAME, shared by every seat: a decision log is a
	// record of a table, and four per-seat files would have to be
	// merged by hand to read one.
	if m.decisionLog != nil {
		dl, err := m.decisionLog.OpenGame(gameID)
		if err != nil {
			// A diagnostic that cannot open its file must not stop a
			// game from being played.
			m.log.Error("bot decision log could not be opened; the game plays without one",
				"game", gameID.String(), "err", err)
		} else {
			bg.dlog = dl
		}
	}
	for _, seat := range seats {
		tier, ok := LookupTier(seat.Tier)
		if !ok {
			m.log.Error("bot seat has an unknown tier; seat will not act",
				"game", gameID.String(), "seat", seat.PlayerID.String(), "tier", seat.Tier)
			continue
		}
		policy, err := factory.NewPolicy(seat)
		if err != nil {
			m.log.Error("bot seat has no policy; seat will not act",
				"game", gameID.String(), "seat", seat.PlayerID.String(), "tier", seat.Tier, "err", err)
			continue
		}
		rcfg := m.configFor(factory, tier)
		if bg.dlog != nil {
			rcfg.Observer = bg.dlog
		}
		r := Start(ctx, room, seat.PlayerID, policy, rcfg, m.bc, m.log.With("game", gameID.String()))
		bg.runners = append(bg.runners, r)
		// The policy NAME is logged next to the tier on purpose. A
		// seat labelled "assisted" that is actually running the
		// random policy is invisible from the outside — the seat
		// plays, it just plays badly — so the one place it can be
		// caught is the line that says which policy the tier built.
		m.log.Info("bot seat runner started",
			"game", gameID.String(), "seat", seat.PlayerID.String(),
			"tier", seat.Tier, "deck", seat.Deck, "policy", policy.Name())
	}
	if len(bg.runners) == 0 {
		cancel()
		bg.finishLog(m.log)
		return
	}
	m.games[gameID] = bg
	// A game that ENDS is never stopped: the runners watch the room,
	// see the state leave StateActive and return on their own. So the
	// log's normal close is here rather than in StopBots — which
	// closes it too, for the game that is deleted or the process that
	// is going away, hence the sync.Once.
	if bg.dlog != nil {
		go func(bg *botGame, log *slog.Logger) {
			for _, r := range bg.runners {
				<-r.Done()
			}
			bg.finishLog(log)
		}(bg, m.log)
	}
	m.log.Info("bot runners started", "game", gameID.String(), "seats", len(bg.runners))
}

// StopBots cancels every runner on the game and waits for them to
// exit. Idempotent.
func (m *Manager) StopBots(gameID uuid.UUID) {
	m.mu.Lock()
	bg, ok := m.games[gameID]
	if ok {
		delete(m.games, gameID)
	}
	m.mu.Unlock()
	if ok {
		bg.cancel()
		for _, r := range bg.runners {
			<-r.Done()
		}
		bg.finishLog(m.log)
		m.log.Info("bot runners stopped", "game", gameID.String())
	}
}

// Runners returns the live runners for a game (tests, admin tools).
func (m *Manager) Runners(gameID uuid.UUID) []*Runner {
	m.mu.Lock()
	defer m.mu.Unlock()
	if bg, ok := m.games[gameID]; ok {
		return append([]*Runner(nil), bg.runners...)
	}
	return nil
}

// Shutdown stops every game's runners.
func (m *Manager) Shutdown() {
	m.mu.Lock()
	ids := make([]uuid.UUID, 0, len(m.games))
	for id := range m.games {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	for _, id := range ids {
		m.StopBots(id)
	}
}
