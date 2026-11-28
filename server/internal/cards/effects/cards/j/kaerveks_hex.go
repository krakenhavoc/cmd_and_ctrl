package cardsj

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaervek's Hex — Sorcery {3}{B} (Mirage, #5633):
//
//	"Kaervek's Hex deals 1 damage to each nonblack creature and an additional 1
//	 damage to each green creature."
//
// A green creature takes the additional 1 whether or not it is black, so a
// black-green creature takes 1 and a nonblack green one takes 2, each in the one
// simultaneous damage event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c69cb2b-14f1-4215-aba8-c107493c3036",
		Name:         "Kaervek's Hex",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OldB794DamageEachCreature(ctx, func(c game.Card) int {
				n := 0
				if !slices.Contains(c.Colors, "B") {
					n++
				}
				if slices.Contains(c.Colors, "G") {
					n++
				}
				return n
			})
		},
	})
}
