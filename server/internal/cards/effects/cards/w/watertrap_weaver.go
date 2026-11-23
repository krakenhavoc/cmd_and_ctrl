package cardsw

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Watertrap Weaver — Creature — Merfolk Wizard {2}{U}, 2/2 (Jumpstart 2022, #5948):
//
//	"When this creature enters, tap target creature an opponent controls.
//	 That creature doesn't untap during its controller's next untap step."
//
// The shared tap-and-freeze (Celestial Regulator's frozen branch): the
// target is tapped and doesn't untap during its controller's next untap
// step (CR 502.3). A target that left or became illegal does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef0ed1b1-d092-46f0-beb2-444e3645d273",
		Name:         "Watertrap Weaver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Watertrap Weaver — tap target creature an opponent controls; it doesn't untap during its controller's next untap step",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Watertrap Weaver — doesn't untap during its controller's next untap step"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
