package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stitched Mangler — Creature — Zombie Horror {2}{U}, 2/3 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"This creature enters tapped.
//	 When this creature enters, tap target creature an opponent controls.
//	 That creature doesn't untap during its controller's next untap step."
//
// An entry replacement (CR 614.1d) and Chill of the Grave's tap-and-freeze
// as an enters trigger: the target is chosen as the trigger goes on the
// stack, and a target that has left or changed controller by resolution is
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c502f1b-7361-48c0-8def-0d3b3e9ff628",
		Name:         "Stitched Mangler",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Stitched Mangler — tap target creature an opponent controls; it doesn't untap during its controller's next untap step",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Stitched Mangler — it doesn't untap during its controller's next untap step"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
