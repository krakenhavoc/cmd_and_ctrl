package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// freerunning.go — freerunning (CR 702.173a, #3172): "You may cast this
// spell for its freerunning cost if a player was dealt combat damage this
// turn by a source that, as it dealt damage, was under your control and
// had assassin creature type or was a commander."
//
// An alternative cost with a gate, shaped exactly like Spectacle: the
// condition is read from the turn tally when the cast is announced and in
// the view, so the offer is not shown to a caster it would be refused
// for, and never read again once the spell is on the stack. The tally
// records WHAT dealt combat damage to which player at the moment it was
// dealt (turn_tally.go), by name, by every subtype the dealer had and,
// since #3172, by whether it was a commander — so an Assassin that died
// in the same combat still counts, and so does a commander you stole.

// Dash is "Dash <cost>" — CR 702.109a: cast the creature for its dash
// cost and it gains haste and returns to its owner's hand at the
// beginning of the next end step. The haste and the return are the
// game-side clause AlternativeCost.Dash (game/dash.go); use this
// constructor, never a hand-built offer.
//
//	AlternativeCosts: []game.AlternativeCost{Dash("{1}{R}")},   // Ragavan
func Dash(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:      "dash",
		Label:    "Dash " + cost,
		ManaCost: cost,
		Dash:     true,
	}
}

// Freerunning is "Freerunning <cost>" for a card whose creature type is
// `creatureType` ("Assassin" on every card printed so far).
//
//	AlternativeCosts: []game.AlternativeCost{Freerunning("{1}{B}", "Assassin")},
func Freerunning(cost, creatureType string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:      "freerunning",
		Label:    "Freerunning " + cost,
		ManaCost: cost,
		Condition: func(g *game.Game, controller uuid.UUID) bool {
			return len(g.PlayersDealtCombatDamageThisTurnBySubtype(controller, creatureType)) > 0 ||
				len(g.PlayersDealtCombatDamageThisTurnByCommander(controller)) > 0
		},
	}
}
