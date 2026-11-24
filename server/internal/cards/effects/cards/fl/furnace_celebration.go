package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Furnace Celebration — Enchantment {1}{R}{R} (Commander Legends):
//
//	"Whenever you sacrifice another permanent, you may pay {2}. If you do,
//	 this enchantment deals 2 damage to any target."
//
// Bloodbriar's trigger (one EventSacrifice per permanent its controller
// sacrifices, the enchantment itself excluded) with Alesha's pay-then-act
// shape: the target is chosen as the trigger goes on the stack, the {2} is
// offered on resolution, and the damage is dealt, to the target if it is
// still legal, only when the mana was paid. The target rides the offer as
// data, so a table waiting on the payer is a restore point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af6d6844-c612-4731-86da-59a8fa02956b",
		Name:         "Furnace Celebration",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.B23YouSacrificedAPermanent(ev, source, lki, g) && ev.CardID != source.InstanceID
			}, furnaceCelebrationLabel,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.MayPay{
							Chooser:  ctx.Controller(),
							Cost:     "{2}",
							Question: furnaceCelebrationLabel + "?",
							Then:     furnaceCelebrationPaid,
							Carry:    []uuid.UUID{t.ID},
						}.Apply(ctx)
					}
					return nil
				}), effects.TargetAny()),
		},
	})
}

const furnaceCelebrationLabel = "Furnace Celebration — pay {2}; if you do, it deals 2 damage to any target"

// furnaceCelebrationPaid is the paid branch. An on-disk key.
var furnaceCelebrationPaid = effects.PayUnlessBranches("pay-unless/furnace-celebration-damage",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) != 1 {
			return nil
		}
		return effects.DealDamage{Source: ctx.Source(), Target: r.Carry[0], Amount: 2}.Apply(ctx)
	}, nil)
