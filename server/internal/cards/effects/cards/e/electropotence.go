package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Electropotence — Enchantment {2}{R} (Zendikar, S82 OLD batch 492, #5583):
//
//	"Whenever a creature you control enters, you may pay {2}{R}. If you do,
//	 that creature deals damage equal to its power to any target."
//
// The any-target is chosen as the trigger goes on the stack; the {2}{R} is
// asked as it resolves (MayPay), and the damage is dealt by the creature that
// entered, with its power read as the payment is answered, or as it last was if
// it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2459c41-82b4-4ca2-b634-11b55ddb63d9",
		Name:         "Electropotence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverAnotherCreatureEntersUnderYourControl(
				"Electropotence — you may pay {2}{R}; if you do, that creature deals damage equal to its power to any target",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.CardID == uuid.Nil {
						return nil
					}
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{R}",
						Question: "Electropotence — pay {2}{R} to have that creature deal damage equal to its power to the target?",
						Then:     electropotencePaid,
						Carry:    []uuid.UUID{item.Trigger.Event.CardID},
					}.Apply(effects.NewContext(g, item))
				}), effects.TargetAny()),
		},
	})
}

var electropotencePaid = effects.PayUnlessBranches("pay-unless/electropotence-damage",
	func(ctx *effects.Context, a game.PayUnlessAnswer) error {
		// Carry is the creature that entered (the asker rides first and is
		// stripped before the branch runs).
		if len(a.Carry) < 1 {
			return nil
		}
		creature := a.Carry[0]
		power := effects.B43PowerNowOrLastKnown(ctx.Game, creature)
		if power <= 0 {
			return nil
		}
		for _, t := range ctx.LegalTargets() {
			return effects.DealDamage{Source: creature, Target: t.ID, Amount: power}.Apply(ctx)
		}
		return nil
	}, nil)
