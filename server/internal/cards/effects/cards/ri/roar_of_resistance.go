package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roar of Resistance — Enchantment {1}{R} (March of the Machine: Phyrexia:
// All Will Be One Commander):
//
//	"Creature tokens you control have haste.
//	 Whenever one or more creatures attack, you may pay {1}{R}. If you do,
//	 creatures attacking your opponents and/or planeswalkers they control get
//	 +2/+0 until end of turn."
//
// The haste grant is a layer 6 static over creature tokens you control
// (creatureTokensYouControl). The trigger watches ANY player's attack
// declaration, once per declaration (OncePerBatch, CR 603.2c): "one or more
// creatures attack" is not limited to yours. The payment is the pay prompt on
// resolution (MayPay, CR 118.12); on payment every creature that is attacking
// one of the controller's opponents, or a planeswalker an opponent controls,
// gets +2/+0. The set is locked as the effect resolves (CR 611.2c), so a
// creature that attacks later in the turn is untouched. Attackers aimed at the
// controller or their planeswalkers are not in it; an opponent's creature
// attacking another opponent is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0753b28-e693-4f5b-a4e4-846e933058d3",
		Name:         "Roar of Resistance",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.CreatureTokensYouControl, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventAttack
			}, "Roar of Resistance — you may pay {1}{R}: creatures attacking your opponents and/or their planeswalkers get +2/+0",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{R}",
						Question: "Roar of Resistance — pay {1}{R} to give creatures attacking your opponents +2/+0?",
						Then:     roarOfResistancePaid,
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}

// roarOfResistancePaid gives +2/+0 to each creature attacking an opponent of
// the ability's controller or a planeswalker an opponent controls. An
// on-disk key: never renamed, never reused.
var roarOfResistancePaid = effects.PayUnlessBranches("pay-unless/roar-of-resistance-pump",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		you := ctx.Controller()
		return effects.BoostUntilEOT{
			Power: 2,
			Match: func(g *game.Game, _ uuid.UUID, c game.Card) bool {
				if !c.IsCreature() || c.AttackingTarget == uuid.Nil {
					return false
				}
				defender := g.DefendingPlayerForAttackForEffect(c.AttackingTarget)
				return defender != uuid.Nil && defender != you
			},
			Label: "Roar of Resistance — +2/+0",
		}.Apply(ctx)
	}, nil)
