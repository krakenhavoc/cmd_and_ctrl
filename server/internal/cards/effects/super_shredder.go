package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Super Shredder — {1}{B} Legendary Creature — Mutant Ninja Human 1/1
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Menace
//	 Whenever another permanent leaves the battlefield, put a +1/+1
//	 counter on Super Shredder."
//
// Any permanent, anyone's, by any route (dies, exile, bounce, a token
// ceasing to exist after it leaves): one trigger per permanent that leaves,
// so a board wipe of five others is five counters. A leaves-the-battlefield
// trigger looks back in time (CR 603.10a), so permanents that leave at the
// same moment as Super Shredder still trigger him, and the counter then has
// nowhere to go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b7ee76bf-d15a-489e-8f05-414788f8f649",
		Name:            "Super Shredder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID != source.InstanceID
			}, "Super Shredder — put a +1/+1 counter on Super Shredder", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
