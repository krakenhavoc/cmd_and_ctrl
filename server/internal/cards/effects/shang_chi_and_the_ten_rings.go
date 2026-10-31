package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shang-Chi and the Ten Rings — {1}{W}{U} Legendary Creature — Human Warrior
// Hero, 2/2 (Marvel Super Heroes, #4455):
//
//	"First strike
//	 Whenever you draw a card, put a +1/+1 counter on Shang-Chi.
//	 When the tenth +1/+1 counter is put on Shang-Chi, you draw five cards
//	 and gain 5 life."
//
// Two printed triggers, two registered ones. The tenth-counter trigger is
// Construct a Cosmic Cube's threshold: the placement event carries the new
// total, and the total before it is read from the log, so a placement that
// carries the count from nine to eleven still puts "the tenth" counter on it,
// and a counter after the tenth does not fire it again (only a crossing does).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5063703b-ddb3-4680-8b16-4cd342dc7d0c",
		Name:            "Shang-Chi and the Ten Rings",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			WheneverYouDraw("Shang-Chi and the Ten Rings — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			On(game.EventCounterPlaced, shangChiTenthCounter,
				"Shang-Chi and the Ten Rings — you draw five cards and gain 5 life",
				Do(DrawCards{N: 5}, GainLife{Amount: 5})),
		},
	})
}

// shangChiTenthCounter is "the tenth +1/+1 counter is put on Shang-Chi".
func shangChiTenthCounter(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Target != source.InstanceID || ev.Label != game.CounterPlusOne || ev.Amount < 10 {
		return false
	}
	return b12CounterTotalBefore(ev, g) < 10
}
