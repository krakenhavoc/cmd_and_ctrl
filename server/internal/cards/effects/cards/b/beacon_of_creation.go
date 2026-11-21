package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beacon of Creation — Sorcery {3}{G} (Fifth Dawn, #5604):
//
//	"Create a 1/1 green Insect creature token for each Forest you
//	 control. Shuffle Beacon of Creation into its owner's library."
//
// The Forests are counted as the spell resolves: anything with the
// Forest subtype, basic or not. Then the Beacon goes into its owner's
// library and that library is shuffled, as the other Beacons do.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55f31c9c-a015-4965-a0c4-788b861dba21",
		Name:         "Beacon of Creation",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			// The shuffle waits for a prompt the tokens may ask (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				forest := ctx.TextSubtype("Forest")
				n := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype(forest) })
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 green Insect"), N: n}.Apply(ctx)
			}),
			effects.Instruction(effects.ShuffleThisSpellIntoItsOwnersLibrary),
		),
	})
}
