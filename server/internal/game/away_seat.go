package game

import "github.com/google/uuid"

// away_seat.go: the table moves on without a player who is gone (ADR
// 0177). The lobby decides when a seat may be moved on and who may do
// it; this file only records what was done and, for a concession, does
// it.
//
// Handing a seat to a bot changes nothing in the rules: the player is
// still in the game and still owns and controls their objects. The bot
// drives the seat from outside, as every bot does. So the only engine
// half of a hand-over, or of taking the seat back, is the log line.

// The EventAwaySeat labels.
const (
	AwayStandIn  = "stand_in"
	AwayConcede  = "concede"
	AwayTakeBack = "take_back"
)

// NoteAwaySeat records a hand-over (AwayStandIn) or a take-back
// (AwayTakeBack) in the log. actor decided (uuid.Nil for the server
// admin), seat is the seat moved, minutes how long it had been offline.
// Returns ErrGameNotActive outside an active game and
// ErrPlayerNotFound for an id that is not a seat.
func (g *Game) NoteAwaySeat(actor, seat uuid.UUID, label string, minutes int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.playerByIDLocked(seat) == nil {
		return ErrPlayerNotFound
	}
	g.EmitEvent(Event{Kind: EventAwaySeat, Actor: actor, Target: seat, Label: label, Amount: minutes})
	return nil
}

// ConcedeForAway concedes seat for a player who is gone (ADR 0177 §5).
// It is a concession (CR 104.3a) through the same path as Concede, so
// CR 800.4a runs as for any player who leaves, and the game ends when
// one player is left (CR 104.2a). Unlike RemoveSeatFromGame the seat is
// not marked Removed: the player did nothing wrong, and may come back
// to watch.
//
// The log says who conceded for whom first, then the concession.
// Returns ErrGameNotActive, ErrPlayerNotFound or ErrPlayerEliminated as
// Concede does.
func (g *Game) ConcedeForAway(actor, seat uuid.UUID, minutes int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(seat)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Eliminated {
		return ErrPlayerEliminated
	}
	g.EmitEvent(Event{Kind: EventAwaySeat, Actor: actor, Target: seat, Label: AwayConcede, Amount: minutes})
	g.concedeLocked(p)
	return nil
}
