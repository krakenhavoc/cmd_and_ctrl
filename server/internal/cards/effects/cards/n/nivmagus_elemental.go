package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nivmagus Elemental — Creature — Elemental {U/R}, 1/2 (Return to Ravnica,
// #5566):
//
//	"Exile an instant or sorcery spell you control: Put two +1/+1 counters on
//	 this creature. (That spell won't resolve.)"
//
// The cost is Synthesis Pod's "Exile a spell you control", narrowed to instants
// and sorceries: the spell leaves the stack as the cost is paid, is not countered
// and does not resolve. The counters go on the Elemental only while it is still
// the permanent the ability came from.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2f9a8a0-509b-4c24-a8cc-f27f77c9eefd",
		Name:         "Nivmagus Elemental",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Exile an instant or sorcery spell you control: Put two +1/+1 counters on this creature.",
			Cost: game.AbilityCost{ExileCards: &game.ExileCost{
				N: 1, Label: "an instant or sorcery spell you control", From: game.ZoneStack,
				Match: func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() },
			}},
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2}),
		}},
	})
}
