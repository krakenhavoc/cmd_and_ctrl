package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Carefree Swinemaster — {2}{G} Creature — Gnome Ranger 1/4
//
//	"Whenever this creature attacks, you may pay {1}{G}. If you do, create a
//	 2/2 green Boar creature token that's tapped and attacking."
//
// The Boar attacks the player or planeswalker the Swinemaster attacked, read
// off the attack event (CR 508.4: it was never declared as an attacker). A
// defender that has left the game by the time the payment is made gets no
// Boar rather than one attacking nobody.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "411cf990-972c-4ee3-9ae4-98602f269407",
		Name:         "Carefree Swinemaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Carefree Swinemaster — you may pay {1}{G}; if you do, create a 2/2 green Boar token that's tapped and attacking",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{G}",
						Question: "Carefree Swinemaster — pay {1}{G} to create a 2/2 green Boar token tapped and attacking?",
						Then:     carefreeSwinemasterPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// carefreeSwinemasterPaid is "if you do, create a 2/2 green Boar creature
// token that's tapped and attacking". An on-disk key: never renamed, never
// reused.
var carefreeSwinemasterPaid = effects.PayUnlessBranches("pay-unless/carefree-swinemaster-boar",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		defender := effects.B17DefendingPlayer(ctx.Game, ctx.Trigger().Event)
		if defender == uuid.Nil {
			return nil
		}
		tmpl := effects.TokenCard("2/2 green Boar")
		tmpl.Tapped = true
		tmpl.AttackingTarget = defender
		return effects.CreateToken{Controller: ctx.Controller(), Template: tmpl, N: 1}.Apply(ctx)
	}, nil)
