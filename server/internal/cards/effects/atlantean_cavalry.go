package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Atlantean Cavalry — Creature — Merfolk Soldier {2}{U}, 3/2 (Marvel
// Super Heroes, #4455):
//
//	"Vigilance
//	 Whenever you draw your second card each turn, put a +1/+1 counter
//	 on this creature."
//
// Knights of Dol Amroth's trigger word for word: b41DrewYourSecondCardThisTurn
// reads the turn's draw tally after the draw, so it fires on the second
// card and no other.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e929b730-9aa2-49e1-9a65-50afc0fbe3e0",
		Name:            "Atlantean Cavalry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Atlantean Cavalry — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
