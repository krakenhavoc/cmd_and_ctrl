package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Digsite Engineer — Creature — Dwarf Artificer {2}{W}, 3/3:
//
//	"Whenever you cast an artifact spell, you may pay {2}. If you do,
//	 create a 0/0 colorless Construct artifact creature token with
//	 "This token gets +1/+1 for each artifact you control.""
//
// A cast trigger (Golem Foundry's YouCast(Artifact())) whose "you may
// pay {2}. If you do" is MayPay, so a pool that can't pay is not asked.
// The token is Urza's Construct, the same printed token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e57025ed-3de6-4b2f-92bd-5481fd33b7f1",
		Name:         "Digsite Engineer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(effects.Artifact()),
				"Digsite Engineer — you may pay {2} to create a Construct token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Digsite Engineer — pay {2} to create a 0/0 Construct artifact creature token?",
						Then:     digsiteEngineerPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// digsiteEngineerPaid is the "if you do". Keyed: never renamed, never reused.
var digsiteEngineerPaid = effects.PayUnlessBranches("pay-unless/digsite-engineer-construct",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Template: effects.UrzaConstructToken(), N: 1}.Apply(ctx)
	}, nil)
