package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oceanus Dragon — Creature — Dragon {4}{U}{U}, 3/5 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Flying
//	 When this creature enters, tap target creature an opponent controls. Goad
//	 it."
//
// The target is picked as the trigger goes on the stack. At resolution a
// target that left or stopped being an opponent's creature is skipped
// (CR 608.2b); otherwise it is tapped and goaded until the controller's next
// turn (GoadAndScheduleClear, CR 701.15).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bf5d44e7-9ff1-4c9a-a62e-628aa5be720d",
		Name:            "Oceanus Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Oceanus Dragon — tap target creature an opponent controls and goad it",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
							return err
						}
						return effects.GoadTarget(g, item)
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
