package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blue Marvel, Adam Brashear — Legendary Creature — Human Scientist Hero
// {3}{U}{U}, 3/5 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Ward {2}
//	 Whenever you draw your second card each turn, put a +1/+1 counter on
//	 Blue Marvel."
//
// Atlantean Cavalry's draw trigger and Ancient Adamantoise's shared Ward
// trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "16f9368a-5da5-4397-a127-c96cdd115b7b",
		Name:            "Blue Marvel, Adam Brashear",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Blue Marvel, Adam Brashear — ward {2}"),
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Blue Marvel, Adam Brashear — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
