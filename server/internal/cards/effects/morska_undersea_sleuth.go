package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morska, Undersea Sleuth — Legendary Creature — Vedalken Fish Detective
// {G}{W}{U}, 2/3 (Murders at Karlov Manor Commander, #5368):
//
//	"You have no maximum hand size.
//	 At the beginning of your upkeep, investigate.
//	 Whenever you draw your second card each turn, put two +1/+1
//	 counters on Morska."
//
// The hand size is Spec.NoMaxHandSize (Kruphix's), the upkeep trigger is
// the ordinary one, and the counters are Atlantean Cavalry's draw tally
// (b41DrewYourSecondCardThisTurn) with two counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "80ea3e65-946d-42c3-a26f-3d74e81d70b4",
		Name:          "Morska, Undersea Sleuth",
		Completeness:  CompletenessFull,
		NoMaxHandSize: true,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Morska, Undersea Sleuth — investigate", Do(Investigate{N: 1})),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Morska, Undersea Sleuth — put two +1/+1 counters on Morska",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
