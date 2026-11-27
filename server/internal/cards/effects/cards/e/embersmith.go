package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embersmith — Creature — Human Artificer {1}{R}, 2/1 (Scars of Mirrodin, S82
// OLD batch 471, #5577):
//
//	"Whenever you cast an artifact spell, you may pay {1}. If you do, this
//	 creature deals 1 damage to any target."
//
// Myrsmith's cast trigger with a target: "any target" is chosen as the trigger
// goes on the stack, the {1} is asked as it resolves (MayPay), and the damage
// is dealt only once it has been paid, to the target if it is still legal
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e16cec8b-e0ac-44ee-91cf-658e204ca120",
		Name:         "Embersmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.Artifact(),
				"Embersmith — you may pay {1} to deal 1 damage to any target",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Embersmith — pay {1} to deal 1 damage to the target?",
						Then:     embersmithPaid,
					}.Apply(effects.NewContext(g, item))
				}), effects.TargetAny()),
		},
	})
}

// embersmithPaid is the paid half: 1 damage from the Embersmith to the
// announced target. An on-disk key: never renamed, never reused.
var embersmithPaid = effects.PayUnlessBranches("pay-unless/s82-old-b471-embersmith-damage",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.B33DamageChosenTargetFromSource(1)(ctx.Game, ctx.Item)
	}, nil)
