package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Erudite Wizard — Creature — Human Wizard {2}{U}, 2/3 (Foundations,
// #4465):
//
//	"Whenever you draw your second card each turn, put a +1/+1 counter on
//	 this creature."
//
// Atlantean Cavalry's trigger. "Each turn" counts draws on any player's
// turn, and only the second draw of a turn fires it, not the third or
// later.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d4b6613-9000-4f53-9737-fe0338f3f629",
		Name:         "Erudite Wizard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Erudite Wizard — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
