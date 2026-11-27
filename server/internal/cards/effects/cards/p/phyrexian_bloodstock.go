package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Bloodstock — Creature — Phyrexian Zombie {4}{B}, 3/3 (Planeshift, S82 OLD b644, #5615):
//
//	"When this creature leaves the battlefield, destroy target white creature. It
//	 can't be regenerated."
//
// A leaves-the-battlefield trigger (CR 603.10a) with a target chosen as it goes on
// the stack; the destruction ignores regeneration shields (CR 701.19c). A target that
// is no longer legal as it resolves is not destroyed (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfeaec62-ca77-46dc-8eae-f627a5e1c02a",
		Name:         "Phyrexian Bloodstock",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisLeaves("Phyrexian Bloodstock — destroy target white creature. It can't be regenerated", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.DestroyTarget{Target: id, CantBeRegenerated: true}.Apply(ctx)
			}), effects.TargetCreature("target white creature", effects.OfColor("W"))),
		},
	})
}
