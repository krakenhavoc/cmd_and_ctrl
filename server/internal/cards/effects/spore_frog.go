package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spore Frog — Creature — Frog {G}, 1/1 (slice 296-m):
//
//	"Sacrifice this creature: Prevent all combat damage that would be
//	 dealt this turn."
//
// Fog's own effect (PreventAllCombatDamageThisTurn, a CR 514.2
// turn-scoped replacement over every combat-damage event) behind a
// sacrifice cost instead of a cast. SacrificeThis pays the cost at
// announce, so the ability is on the stack independent of the Frog,
// which is already in the graveyard by the time it resolves — exactly
// as printed, and exactly why the effect is a scoped rule rather than
// anything read off the source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97db6c39-e690-49b6-93a6-e51b8dfad10b",
		Name:         "Spore Frog",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Prevent all combat damage that would be dealt this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PreventAllCombatDamageThisTurn{
					Label: "Spore Frog: prevent combat damage",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
