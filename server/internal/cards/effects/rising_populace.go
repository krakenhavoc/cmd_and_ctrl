package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rising Populace — Creature — Human {2}{W}, 2/2 (Ravnica Remastered):
//
//	"Whenever another creature or planeswalker you control dies, put a
//	 +1/+1 counter on this creature."
//
// Same trigger shape as Edgar, Ancient Bloodlord: what the permanent WAS as
// it died (CR 603.10a), and never the Populace itself. The counter goes on
// the Populace only while it is still on the battlefield (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "340de754-9d64-4824-936f-643be4a939db",
		Name:         "Rising Populace",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return gapB77AnotherCreatureOrPlaneswalkerYouControlDied(ev, source, g)
			},
			Key:     "Rising Populace — put a +1/+1 counter on this creature",
			Purpose: game.Purpose{DeathPayoff: true},
			Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
