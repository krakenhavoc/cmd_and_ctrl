package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kor Hookmaster — {2}{W} Creature — Kor Soldier 2/2 (Eternal Masters, S82
// OLD, #5542):
//
//	"When this creature enters, tap target creature an opponent controls. That
//	 creature doesn't untap during its controller's next untap step."
//
// Spire Patrol's enters trigger: the tap and the one-shot freeze (TapAndFreeze,
// the ADR 0058 mark) on a creature an opponent controls chosen as the trigger
// goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6941577-8262-42ef-bf32-f6505d90b73a",
		Name:         "Kor Hookmaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Kor Hookmaster — tap target creature an opponent controls; it doesn't untap during its controller's next untap step",
					func(g *game.Game, item *game.StackItem) error {
						return effects.B751TapFreezeTargets(effects.NewContext(g, item), uuid.Nil, "Kor Hookmaster — doesn't untap during its controller's next untap step")
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
