package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jedit Ojanen, Mercenary — Legendary Creature — Cat Mercenary {1}{W}{U},
// 3/3 (Dominaria United Commander, S82):
//
//	"Whenever Jedit Ojanen or another legendary creature you control
//	 enters, you may pay {G}. If you do, create a 2/2 green Cat Warrior
//	 creature token with forestwalk."
//
// One trigger over the shared this-or-another-legendary condition; the
// payment is the may-pay prompt and the token is its paid branch.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d80fa33-9032-4c16-8355-2f18a71af67c",
		Name:         "Jedit Ojanen, Mercenary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.ThisOrAnotherLegendaryCreatureYouControlEntered,
				"Jedit Ojanen, Mercenary — you may pay {G} to create a 2/2 Cat Warrior with forestwalk",
				jeditOjanenOffer),
		},
	})
}

func jeditOjanenOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{G}",
		Question: "Jedit Ojanen, Mercenary — pay {G} to create a 2/2 Cat Warrior with forestwalk?",
		Then:     jeditOjanenPaid,
	}.Apply(effects.NewContext(g, item))
}

var jeditOjanenPaid = effects.PayUnlessBranches("pay-unless/jedit-ojanen-mercenary-cat-warrior",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Template: effects.TokenCard("2/2 green Cat Warrior with forestwalk"), N: 1}.Apply(ctx)
	}, nil)
