package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Ironfoot — Snow Artifact Creature — Phyrexian Construct {3},
// 3/4 (Jumpstart 2022):
//
//	"This creature doesn't untap during your untap step.
//	 {1}{S}: Untap this creature. ({S} can be paid with one mana from a snow
//	 source.)"
//
// Brass Man's untap-step restriction and Soliton's untap ability at
// {1}{S}, the {S} paid only by mana a snow source made (CR 107.4h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "83b4acd0-751d-4882-a32d-6c089bc652ee",
		Name:                  "Phyrexian Ironfoot",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Activated: []effects.ActivatedAbility{{
			UntapSelf: true,
			Label:     "{1}{S}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{1}{S}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
