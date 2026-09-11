package aiseat

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// Tier names. Only TierRandom has a policy today; the others arrive
// with sub-PRs 6 and 7 and are not offered until they do.
const (
	TierRandom = "random"
)

// Manager owns the runners for every bot seat on the server: one
// Start per game (from Lobby.Start) and one Stop per game (from
// Lobby.Delete). Satisfies lobby.BotHost. Added in S31 sub-PR 4.
type Manager struct {
	bc  Broadcaster
	cfg Config
	log *slog.Logger

	mu    sync.Mutex
	games map[uuid.UUID]*botGame
}

type botGame struct {
	cancel  context.CancelFunc
	runners []*Runner
}

// NewManager builds a Manager. bc is the hub (may be nil in tests);
// cfg is the pacing every runner gets — DefaultConfig in production.
func NewManager(bc Broadcaster, cfg Config, log *slog.Logger) *Manager {
	if log == nil {
		log = slog.Default()
	}
	return &Manager{bc: bc, cfg: cfg, log: log, games: make(map[uuid.UUID]*botGame)}
}

// Tiers lists the policy tiers a bot may be added with.
func (m *Manager) Tiers() []string { return []string{TierRandom} }

// PolicyFor builds a fresh policy for a tier. Unknown tiers get the
// random policy with a warning rather than an error: by the time a
// seat starts, the tier was validated at add time, so this is a
// belt-and-braces path, not a request path.
func (m *Manager) PolicyFor(tier string) Policy {
	switch tier {
	case TierRandom:
		return NewRandomPolicy(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	default:
		m.log.Warn("unknown bot tier; using random", "tier", tier)
		return NewRandomPolicy(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
}

// StartBots launches a runner per seat on the room. Calling it twice
// for the same game replaces the first set (the old runners are
// cancelled), so a restart cannot double-drive a seat.
func (m *Manager) StartBots(room *ws.Room, seats []lobby.BotSeat) {
	if room == nil || len(seats) == 0 {
		return
	}
	gameID := room.Game.ID
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.games[gameID]; ok {
		old.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	bg := &botGame{cancel: cancel}
	for _, seat := range seats {
		r := Start(ctx, room, seat.PlayerID, m.PolicyFor(seat.Tier), m.cfg, m.bc, m.log.With("game", gameID.String()))
		bg.runners = append(bg.runners, r)
	}
	m.games[gameID] = bg
	m.log.Info("bot runners started", "game", gameID.String(), "seats", len(seats))
}

// StopBots cancels every runner on the game. Idempotent.
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
