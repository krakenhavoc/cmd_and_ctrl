package game

import "github.com/google/uuid"

// SeatDriver reports who plays the seated player id: a bot runner, an
// AI agent through an MCP client, or (both false) a person. Both are
// false for an id with no seat. Takes the read lock, so the room can
// label a commit (ADR 0123 §3's seat_kind) without reading the flags
// unguarded.
func (g *Game) SeatDriver(id uuid.UUID) (bot, agent bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	p := g.playerByIDLocked(id)
	if p == nil {
		return false, false
	}
	return p.IsBot, p.Agent
}
