package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imperial Cosmographer — Creature — Kree Scientist {2}{U}, 2/3 (Marvel
// Super Heroes Commander, #4455):
//
//	"Vigilance
//	 Whenever another creature you control leaves the battlefield without
//	 dying, put two +1/+1 counters on this creature."
//
// Dour Port-Mage's b13OtherCreatureYouControlLeftWithoutDying, but per
// creature (this wording is not "one or more"), so there is no
// OncePerBatch: two creatures bounced together give two triggers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b2aa052c-08b7-4334-9ea7-e78da57ff9ea",
		Name:            "Imperial Cosmographer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13OtherCreatureYouControlLeftWithoutDying(ev, source, g)
			}, "Imperial Cosmographer — put two +1/+1 counters on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
