package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sleeper Dart — Artifact {2} (Ikoria: Lair of Behemoths, #5512):
//
//	"When this artifact enters, draw a card.
//	 {T}, Sacrifice this artifact: Target creature doesn't untap during its
//	 controller's next untap step."
//
// The target is chosen as the ability goes on the stack and the artifact is
// sacrificed as a cost, so the freeze outlives it. The creature is not tapped
// by this: it only stays tapped through its controller's next untap step, and
// an untapped creature is unaffected in effect. A target that has left or is
// no longer legal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f1b6130-3f18-4989-922b-438e41d683da",
		Name:         "Sleeper Dart",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sleeper Dart — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Target creature doesn't untap during its controller's next untap step.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.DoesntUntapNextUntapStep{Targets: []uuid.UUID{id}, Label: "Sleeper Dart — doesn't untap during its controller's next untap step"}.Apply(ctx)
			},
		}},
	})
}
