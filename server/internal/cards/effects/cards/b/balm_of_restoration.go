package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balm of Restoration — Artifact {2} (Fallen Empires, S82 OLD b855, #5637):
//
//	"{1}, {T}, Sacrifice this artifact: Choose one —
//	 • You gain 2 life.
//	 • Prevent the next 2 damage that would be dealt to any target this turn."
//
// The mode, and the second mode's target, are chosen as the ability is activated
// (CR 602.2b, 601.2b-c); the sacrifice is a cost. The shield is a charged prevention
// of 2 on the target (CR 615.7), as Pearl Shard's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45a40b4c-1d37-4d87-adf1-abf0ea4c86b4",
		Name:         "Balm of Restoration",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Choose one — You gain 2 life; or prevent the next 2 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Modes: effects.ChooseOne(
				effects.ModeDoing("You gain 2 life.", nil,
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
					}),
				effects.ModeDoing("Prevent the next 2 damage that would be dealt to any target this turn.", effects.TargetAny(),
					func(item *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.PreventNextDamageToTheTarget("Balm of Restoration", 2)(ctx.Game, item)
					}),
			),
		}},
	})
}
