package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dromar, the Banisher — Legendary Creature — Dragon {3}{W}{U}{B}, 6/6 (Invasion, S82 #5616):
//
//	"Flying
//	 Whenever Dromar deals combat damage to a player, you may pay {2}{U}. If you
//	 do, choose a color, then return all creatures of that color to their owners'
//	 hands."
//
// Rith, the Awakener's shape: the payment is asked as the ability resolves
// (MayPay); the paid branch names the colour and a keyed tail bounces every
// creature of it, whoever controls it, Dromar included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "35ad7f82-a1b1-4aa7-a8b0-3f99ca2cff52",
		Name:            "Dromar, the Banisher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Dromar, the Banisher — you may pay {2}{U}; if you do, choose a color and return all creatures of that color to their owners' hands",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{U}",
						Question: "Dromar, the Banisher — pay {2}{U}?",
						Then:     dromarPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var dromarPaid = effects.PayUnlessBranches("pay-unless/dromar-the-banisher-bounce",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		effects.ChooseColorTail(game.ColorForBenefit, ctx.Game, ctx.Controller(), ctx.Source(),
			"Dromar, the Banisher — choose a color", ctx.ItemTail(dromarBounce))
		return nil
	}, nil)

var dromarBounce = effects.CardTail("then/dromar-the-banisher-bounce", func(ctx *effects.Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" {
		return nil
	}
	return effects.BounceAllMatching{Match: effects.And(effects.Creature(), effects.OfColor(color))}.Apply(ctx)
})
