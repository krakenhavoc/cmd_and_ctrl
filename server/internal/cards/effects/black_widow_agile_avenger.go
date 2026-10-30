package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Widow, Agile Avenger — Legendary Creature — Human Spy Hero {1}{R}{W},
// 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Menace
//	 Whenever an opponent draws their second card each turn, put a +1/+1
//	 counter on Black Widow and you draw a card."
//
// Faerie Mastermind's trigger word for word (opponentDrewTheirSecondCardThisTurn
// reads the turn's draw tally, so it fires on the second card and no other),
// with the counter first and the draw second.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "df7abfb0-21df-4359-be81-0f3c09e1425e",
		Name:            "Black Widow, Agile Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentDrewTheirSecondCardThisTurn(ev, source, g)
			}, "Black Widow, Agile Avenger — put a +1/+1 counter on this creature and draw a card",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, DrawCards{N: 1})),
		},
	})
}
