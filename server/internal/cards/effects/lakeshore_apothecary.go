package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lakeshore Apothecary — Creature — Human Cleric {1}{U}, 1/2:
//
//	"Vigilance
//	 Whenever you draw your second card each turn, put a +1/+1
//	 counter on this creature."
//
// Knights of Dol Amroth's trigger on a vigilant body; see that file
// for how "your second card" is counted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "80bcb154-b4be-4bbd-bbc1-b116a59045b7",
		Name:            "Lakeshore Apothecary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Lakeshore Apothecary — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
