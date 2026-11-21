package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rith, the Awakener — Legendary Creature — Dragon {3}{R}{G}{W}, 6/6
// (Dominaria Remastered, #5948):
//
//	"Flying
//	 Whenever Rith deals combat damage to a player, you may pay {2}{G}. If
//	 you do, choose a color, then create a 1/1 green Saproling creature
//	 token for each permanent of that color."
//
// The payment is asked as the ability resolves (CR 603.5, MayPay); "if you
// do" is its paid branch, which asks for the colour and hands the answer to
// a keyed tail. The tokens are counted when the tail runs, after the
// colour is named: every permanent of that colour, whoever controls it,
// the Rith that dealt the damage included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "97e5d788-2dd3-4c8e-9ab6-f4ea9c7a9a3d",
		Name:            "Rith, the Awakener",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtCombatDamageToAPlayer,
				"Rith, the Awakener — you may pay {2}{G}; if you do, choose a color and create a Saproling for each permanent of that color",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{G}",
						Question: "Rith, the Awakener — pay {2}{G}?",
						Then:     rithPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

var rithPaid = PayUnlessBranches("pay-unless/rith-the-awakener-saprolings",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		ChooseColorTail(game.ColorForBenefit, ctx.Game, ctx.Controller(), ctx.Source(),
			"Rith, the Awakener — choose a color", ctx.ItemTail(rithSaprolings))
		return nil
	}, nil)

var rithSaprolings = CardTail("then/rith-the-awakener-saprolings", func(ctx *Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" {
		return nil
	}
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.HasColor(color) {
			n++
		}
	}
	if n <= 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 green Saproling"), N: n}.Apply(ctx)
})
