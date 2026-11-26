package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golden Urn — Artifact {1} (Scars of Mirrodin, S82 OLD batch 470, #5577):
//
//	"At the beginning of your upkeep, you may put a charge counter on this
//	 artifact.
//	 {T}, Sacrifice this artifact: You gain life equal to the number of charge
//	 counters on this artifact."
//
// Barrin's Codex's optional upkeep counter, and its sacrifice-for-X: the Urn
// is sacrificed as a cost, so the life is counted from it as it last existed
// on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	const charge = game.CounterCharge
	effects.Register(effects.Spec{
		OracleID:     "ddb19cac-20b0-4ab8-a31d-c5c155491780",
		Name:         "Golden Urn",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Golden Urn — put a charge counter on this artifact", effects.PutACounterOnThis(charge)),
				"Golden Urn — put a charge counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: You gain life equal to the number of charge counters on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Player: ctx.Controller(), Amount: effects.CountersOnThisAsItResolves(charge)(ctx)}.Apply(ctx)
			})),
		}},
	})
}
