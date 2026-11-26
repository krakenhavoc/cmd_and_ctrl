package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wavecrash Triton — Creature — Merfolk Wizard {2}{U}, 1/4 (Theros, S82 OLD
// b388):
//
//	"Heroic — Whenever you cast a spell that targets this creature, tap target
//	 creature an opponent controls. That creature doesn't untap during its
//	 controller's next untap step."
//
// Berg Strider's tap and freeze (TapAndFreeze), with the target chosen as the
// trigger goes on the stack (CR 603.3d) from the creatures an opponent
// controls. With no such creature the trigger is removed from the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a7e68ed-156c-4680-851b-ae8de4a808a5",
		Name:         "Wavecrash Triton",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Heroic("Wavecrash Triton — tap target creature an opponent controls; it doesn't untap during its controller's next untap step (heroic)",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Wavecrash Triton — it doesn't untap during its controller's next untap step"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
