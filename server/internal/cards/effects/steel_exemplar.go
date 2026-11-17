package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steel Exemplar — Artifact Creature — Wizard {5}, 4/4 (The Brothers' War):
//
//	"Trample
//	 This creature enters with two +1/+1 counters on it unless two or more
//	 colors of mana were spent to cast it."
//
// A CR 614.1c self-replacement over the entering spell's payment
// (SelfEntersWithCountersFromSpentMana). An entry that was not a cast
// spent no colours, so a reanimated Exemplar gets its counters.
func init() {
	Register(Spec{
		OracleID:        "ab3b72e3-9a49-4c95-8016-22fa02ad7663",
		Name:            "Steel Exemplar",
		Completeness:    CompletenessFull,
		ReadsManaSpent:  game.ReadsOwnPayment,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			SelfEntersWithCountersFromSpentMana(game.CounterPlusOne,
				"Enters with two +1/+1 counters on it unless two or more colors of mana were spent to cast it.",
				func(m game.ManaSpent) int {
					if m.ColorCount() >= 2 {
						return 0
					}
					return 2
				}),
		},
	})
}
