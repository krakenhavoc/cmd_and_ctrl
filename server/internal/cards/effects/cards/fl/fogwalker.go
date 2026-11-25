package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fogwalker — Creature — Spirit {1}{U}, 1/3 (Eldritch Moon, S82 OLD,
// #5541):
//
//	"Skulk (This creature can't be blocked by creatures with greater
//	 power.)
//	 When this creature enters, target creature an opponent controls doesn't
//	 untap during its controller's next untap step."
//
// Berg Strider's freeze without the tap: the target is chosen as the trigger
// goes on the stack and held through its controller's next untap step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c597cfae-3f92-4d31-8bc1-5a2da995f1bf",
		Name:            "Fogwalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"skulk"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Fogwalker — target creature an opponent controls doesn't untap during its controller's next untap step",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.DoesntUntapNextUntapStep{Targets: []uuid.UUID{id}, Label: "Fogwalker — doesn't untap during its controller's next untap step"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
