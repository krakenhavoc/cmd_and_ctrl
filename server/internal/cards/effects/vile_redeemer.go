package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vile Redeemer — Creature — Eldrazi {2}{G}, 3/3:
//
//	"Devoid (This card has no color.)
//	 Flash
//	 When you cast this spell, you may pay {C}. If you do, create a 1/1
//	 colorless Eldrazi Scion creature token for each nontoken creature that
//	 died under your control this turn. Those tokens have "Sacrifice this
//	 token: Add {C}.""
//
// The cast trigger (Abundant Maw's WhenYouCastThisSpell) offers a {C}
// payment as it resolves (colourless mana only, CR 107.4c); when it is paid
// the nontoken creatures that died under the controller's control this turn
// are counted then (gapB44NontokenCreaturesDiedUnder) and as many Eldrazi
// Scions are made. The Scion is the printed token with its own mana
// ability, so the ability the text gives is the token's.
func init() {
	const name = "Vile Redeemer"
	Register(Spec{
		OracleID:        "a7a88bcc-eac2-4f74-804f-b49d16560184",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flash"},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell(name+" — you may pay {C}; if you do, create an Eldrazi Scion for each nontoken creature that died under your control this turn",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{C}",
						Question: name + " — pay {C} to create an Eldrazi Scion for each nontoken creature that died under your control this turn?",
						Then:     vileRedeemerPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// vileRedeemerPaid creates the Scions once {C} is paid. An on-disk key.
var vileRedeemerPaid = PayUnlessBranches("pay-unless/vile-redeemer-scions", func(ctx *Context, r game.PayUnlessAnswer) error {
	n := gapB44NontokenCreaturesDiedUnder(ctx.Game, ctx.Controller())
	if n <= 0 {
		return nil
	}
	return CreateToken{Template: b28EldraziScionToken(), N: n}.Apply(ctx)
}, nil)
