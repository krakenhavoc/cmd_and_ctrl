package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ticket Turbotubes — Artifact {3}:
//
//	"{T}: Add one mana of any color.
//	 {3}, {T}: You get {TK} (a ticket counter)."
//
// The second ability adds no mana, so it is an ordinary CR 602 activation
// on the stack; the ticket is a player counter (CR 107.17, ADR 0208).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85abdc70-6fab-4d44-b286-b044eb839c5e",
		Name:         "Ticket Turbotubes",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: You get {TK}.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.Do(effects.GetTickets{N: 1}),
		}},
	})
}
