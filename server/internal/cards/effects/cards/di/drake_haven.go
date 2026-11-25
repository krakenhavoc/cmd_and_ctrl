package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drake Haven — {2}{U} Enchantment (Commander 2020):
//
//	"Whenever you cycle or discard a card, you may pay {1}. If you do, create
//	 a 2/2 blue Drake creature token with flying."
//
// Cycling is discarding the card as a cost (CR 702.29a), so every cycled card
// is also a discarded one and the one event, EventDiscardCard, covers both
// words: a card that is cycled triggers it once, not twice. The payment is
// asked as the trigger resolves (Lunar Mystic's MayPay) and the token is made
// only if it is paid.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86d3def3-53fa-4cad-bb25-d3275af1e3f5",
		Name:         "Drake Haven",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Drake Haven — you may pay {1} to create a 2/2 blue Drake creature token with flying",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Drake Haven — pay {1} to create a 2/2 blue Drake creature token with flying?",
						Then:     drakeHavenPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var drakeHavenPaid = effects.PayUnlessBranches("pay-unless/drake-haven-drake", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	return effects.CreateToken{Template: effects.TokenCard("2/2 blue Drake with flying"), N: 1}.Apply(ctx)
}, nil)
