package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Weldfast Engineer — Creature — Human Artificer {1}{B}{R}, 3/3 (Aether Revolt,
// S82 older sets, #5536):
//
//	"At the beginning of combat on your turn, target artifact creature you
//	 control gets +2/+0 until end of turn."
//
// The target is chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74adb283-9bef-4a6f-9489-bcfcba53e8d7",
		Name:         "Weldfast Engineer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat("Weldfast Engineer — target artifact creature you control gets +2/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: 2, Label: "Weldfast Engineer — +2/+0"}.Apply(ctx)
				}),
				effects.TargetCreature("target artifact creature you control", effects.Artifact(), effects.YouControl())),
		},
	})
}
