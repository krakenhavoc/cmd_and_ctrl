package aiseat

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// SeatSpec is one bot seat to run: which seat, and at what tier. The
// lobby builds these from its own seat list and hands them to
// StartBots; keeping the type here rather than in lobby is what lets
// aiseat stay the lower layer, imported by the lobby and importing
// nothing of it.
type SeatSpec struct {
	PlayerID uuid.UUID
	Tier     string
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

	mu    sync.Mutex
	games map[uuid.UUID]*botGame
}

type botGame struct {
	cancel  context.CancelFunc
	runners []*Runner
}

// NewManager builds a Manager that paces each runner by its tier
// (ConfigFor). bc is the hub and may be nil in tests.
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

// Tiers lists the policy tiers a bot may be added with. Only the
// available ones — the lobby validates an add request against this,
// and an unavailable tier must be refused rather than downgraded.
func (m *Manager) Tiers() []string {
	avail := AvailableTiers()
	out := make([]string, 0, len(avail))
	for _, t := range avail {
		out = append(out, string(t))
	}
	return out
}

func (m *Manager) configFor(t Tier) Config {
	if m.cfg != nil {
		return *m.cfg
	}
	return ConfigFor(t)
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
	}
	ctx, cancel := context.WithCancel(context.Background())
	bg := &botGame{cancel: cancel}
	for _, seat := range seats {
		tier, ok := LookupTier(seat.Tier)
		if !ok {
			m.log.Error("bot seat has an unknown tier; seat will not act",
				"game", gameID.String(), "seat", seat.PlayerID.String(), "tier", seat.Tier)
			continue
		}
		policy, err := NewPolicy(tier)
		if err != nil {
			m.log.Error("bot seat has no policy; seat will not act",
				"game", gameID.String(), "seat", seat.PlayerID.String(), "tier", seat.Tier, "err", err)
			continue
		}
		r := Start(ctx, room, seat.PlayerID, policy, m.configFor(tier), m.bc, m.log.With("game", gameID.String()))
		bg.runners = append(bg.runners, r)
	}
	if len(bg.runners) == 0 {
		cancel()
		return
	}
	m.games[gameID] = bg
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
