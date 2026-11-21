package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frost Lynx — Creature — Elemental Cat {2}{U}, 2/2 (Magic 2015, #5556):
//
//	"When this creature enters, tap target creature an opponent controls.
//	 That creature doesn't untap during its controller's next untap step."
//
// White Dragon's Cold Breath on a ground body: the shared tap-and-freeze
// (b751TapFreezeTargets) on the trigger's target, which is chosen as the
// trigger goes on the stack and re-checked as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9400e348-b0c6-4aa3-b92c-1fbc8bc3baf9",
		Name:         "Frost Lynx",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Frost Lynx — tap target creature an opponent controls; it doesn't untap during its controller's next untap step", func(g *game.Game, item *game.StackItem) error {
				return effects.B751TapFreezeTargets(effects.NewContext(g, item), uuid.Nil, "Frost Lynx")
			}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
