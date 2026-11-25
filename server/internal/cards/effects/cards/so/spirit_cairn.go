package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spirit Cairn — {2}{W} Enchantment (Commander 2020):
//
//	"Whenever a player discards a card, you may pay {W}. If you do, create a
//	 1/1 white Spirit creature token with flying."
//
// Drake Haven's shape over every player's discards (a cycled card is
// discarded too, CR 702.29a); the Cairn's controller pays and gets the token.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae896d87-59c7-4d8d-a137-9e471d3e174e",
		Name:         "Spirit Cairn",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventDiscardCard
			}, "Spirit Cairn — you may pay {W} to create a 1/1 white Spirit creature token with flying",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{W}",
						Question: "Spirit Cairn — pay {W} to create a 1/1 white Spirit creature token with flying?",
						Then:     spiritCairnPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var spiritCairnPaid = effects.PayUnlessBranches("pay-unless/spirit-cairn-spirit", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	return effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1}.Apply(ctx)
}, nil)
