package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wanted Scoundrels — Creature — Human Pirate {1}{B}, 4/3 (Ixalan, S82 older
// sets):
//
//	"When this creature dies, target opponent creates two Treasure tokens."
//
// Hunted Lammasu's "target opponent creates": the opponent is the target
// (CR 603.3d), re-checked as the trigger resolves, and the Treasures are
// created by them, so they own and control them (CR 111.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c63c4ca-5d5a-4ab9-af0d-ed1254b581cc",
		Name:         "Wanted Scoundrels",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Wanted Scoundrels — target opponent creates two Treasure tokens",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return effects.CreateToken{Controller: t.ID, Template: effects.TreasureToken(), N: 2}.Apply(ctx)
						}
					}
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
