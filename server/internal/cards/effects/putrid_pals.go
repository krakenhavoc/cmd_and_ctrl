package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Putrid Pals — {2}{B/G}{B/G} Creature — Human Ooze Mutant 3/3 (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"Deathtouch
//	 Disappear — This creature enters with two +1/+1 counters on it if a
//	 permanent left the battlefield under your control this turn."
//
// Disappear is an ability word (CR 207.2c) with revolt's condition, so this
// is Narnam Renegade's self-replacement (CR 614.1c) with two counters, read
// as the creature enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3aa3b908-e8d3-4944-8ced-5219d4129594",
		Name:            "Putrid Pals",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Replacements:    []game.ReplacementEffect{SelfEntersWithCountersIfRevolt(game.CounterPlusOne, 2)},
	})
}
