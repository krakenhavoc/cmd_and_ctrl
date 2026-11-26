package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Knowledge and Power — Enchantment {4}{R} (Journey into Nyx, S82 OLD, #5558):
//
//	"Whenever you scry, you may pay {2}. If you do, this enchantment deals 2
//	 damage to any target."
//
// Furnace Celebration's targeted "you may pay {2}" over a watch on EventScry
// (emitted once a scry is complete, and not for a scry of an empty library,
// which did nothing). The target is chosen as the trigger is put on the
// stack, so the payment is the only choice left at resolution; a target that
// is gone then takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed68b5ca-c52a-40e0-89ac-6276a3c4fdeb",
		Name:         "Knowledge and Power",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventScry, effects.ByYou, knowledgeAndPowerLabel,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.MayPay{
							Chooser:  ctx.Controller(),
							Cost:     "{2}",
							Question: knowledgeAndPowerLabel + "?",
							Then:     knowledgeAndPowerPaid,
							Carry:    []uuid.UUID{t.ID},
						}.Apply(ctx)
					}
					return nil
				}), effects.TargetAny()),
		},
	})
}

const knowledgeAndPowerLabel = "Knowledge and Power — pay {2}; if you do, it deals 2 damage to any target"

// knowledgeAndPowerPaid is the paid branch. An on-disk key.
var knowledgeAndPowerPaid = effects.PayUnlessBranches("pay-unless/s82-old-b369-knowledge-and-power-damage",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) != 1 {
			return nil
		}
		return effects.DealDamage{Source: ctx.Source(), Target: r.Carry[0], Amount: 2}.Apply(ctx)
	}, nil)
