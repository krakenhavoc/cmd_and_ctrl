package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grovetender Druids — Creature — Elf Druid Ally {2}{G}{W}, 3/3 (Battle for
// Zendikar, #5547):
//
//	"Rally — Whenever this creature or another Ally you control enters, you
//	 may pay {1}. If you do, create a 1/1 green Plant creature token."
//
// Rally is one trigger per Ally entering under your control, this creature
// included. The payment is the may-pay prompt, and the token is made only
// when the mana was really paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1fc8b186-bb74-4fe1-a2da-e8ce3074f988",
		Name:         "Grovetender Druids",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB313Rally("Grovetender Druids — you may pay {1}; if you do, create a 1/1 green Plant creature token", grovetenderDruidsOffer),
		},
	})
}

func grovetenderDruidsOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Grovetender Druids — pay {1} to create a 1/1 green Plant creature token?",
		Then:     grovetenderDruidsPaid,
	}.Apply(effects.NewContext(g, item))
}

var grovetenderDruidsPaid = effects.PayUnlessBranches("pay-unless/grovetender-druids-plant",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Template: effects.TokenCard("1/1 green Plant"), N: 1}.Apply(ctx)
	}, nil)
