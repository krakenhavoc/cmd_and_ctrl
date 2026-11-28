package cardsto

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tropical Storm — Sorcery {X}{G} (Mirage, #5633):
//
//	"Tropical Storm deals X damage to each creature with flying and 1 additional
//	 damage to each blue creature."
//
// A blue creature takes the additional 1 whether or not it flies, in the one
// simultaneous damage event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b3b0bef-cdbb-41c2-b8f1-ddda9903b43a",
		Name:         "Tropical Storm",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			return effects.OldB794DamageEachCreature(ctx, func(c game.Card) int {
				n := 0
				if game.HasKeyword(&c, "flying") {
					n += x
				}
				if slices.Contains(c.Colors, "U") {
					n++
				}
				return n
			})
		},
	})
}
