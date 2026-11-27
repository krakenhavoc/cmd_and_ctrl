package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Searing Meditation — Enchantment {1}{R}{W} (Ravnica: City of Guilds):
//
//	"Whenever you gain life, you may pay {2}. If you do, this enchantment deals 2
//	 damage to any target."
//
// Cradle of Vitality's shape: one trigger per life-gain event (lifelink
// included), with its any-target picked as it goes on the stack (CR 603.3d).
// The {2} is offered as the ability resolves, and the damage comes from the
// enchantment by last-known information if it has left (CR 608.2h). A target
// that left or became illegal gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6342d2c-06e3-4443-9276-b49f4a06a4da",
		Name:         "Searing Meditation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouGainLife("Searing Meditation — you may pay {2}; if you do, it deals 2 damage to any target",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Searing Meditation — pay {2} to deal 2 damage to the target?",
						Then:     searingMeditationPaid,
					}.Apply(effects.NewContext(g, item))
				}),
				effects.TargetAny()),
		},
	})
}

// searingMeditationPaid is the paid half, keyed: the asking trigger and its
// target ride the prompt.
var searingMeditationPaid = effects.PayUnlessBranches("pay-unless/searing-meditation-damage",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		t, ok := effects.FirstLegalTarget(ctx)
		if !ok {
			return nil
		}
		return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
	}, nil)
