package game

import "github.com/google/uuid"

// faced_dilemma.go — #4973 (Star Trek): "you face a dilemma as you
// choose one or more modes for a spell or ability" (Jett Reno, Problem
// Solver). The Comprehensive Rules have no entry for it; the reminder
// text defines it as the act of choosing modes (CR 700.2a for a spell
// or activated ability, CR 700.2b for a triggered one). The engine
// already has one place for each of those choices, and each emits
// EventFacedDilemma once the choice is made.
//
// Also the extra-turn count H.M.S. Bounty reads, which shipped in the
// same batch: Player.ExtraTurnsBegun.

// emitFacedDilemmaLocked records that `player` chose one or more modes
// for the spell or ability `source` (the spell, or the ability's
// source).
//
// Caller must hold g.mu in write mode.
func (g *Game) emitFacedDilemmaLocked(player, source uuid.UUID) {
	if player == uuid.Nil {
		return
	}
	g.EmitEvent(Event{
		Kind:   EventFacedDilemma,
		Actor:  player,
		Source: source,
		CardID: source,
	})
}

// ExtraTurnsBegunForEffect is how many extra turns `player` has begun
// this game (Player.ExtraTurnsBegun): "if you … haven't begun an extra
// turn this game" is a zero here. A player not in the game reads 0.
//
// Caller must hold g.mu.
func (g *Game) ExtraTurnsBegunForEffect(player uuid.UUID) int {
	if p := g.playerByIDLocked(player); p != nil {
		return p.ExtraTurnsBegun
	}
	return 0
}
