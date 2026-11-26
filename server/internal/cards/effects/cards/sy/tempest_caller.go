package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tempest Caller — Creature — Merfolk Wizard {2}{U}{U}, 2/3 (Ixalan, S82
// older sets):
//
//	"When this creature enters, tap all creatures target opponent controls."
//
// The opponent is the target (CR 603.3d), re-checked as the trigger resolves;
// every creature they control is tapped (Assassin Gauntlet's second half).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6ec5a5c-144d-4ac3-bd05-2dc360c670fd",
		Name:         "Tempest Caller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Tempest Caller — tap all creatures target opponent controls",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
							return nil
						}
						return effects.B29TapAllCreaturesControlledBy(ctx, item.Targets[0].ID)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
