package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Narnam Renegade — Creature — Elf Warrior {G}, 1/2:
//
//	"Deathtouch
//	 Revolt — This creature enters with a +1/+1 counter on it if a
//	 permanent left the battlefield under your control this turn."
//
// A CR 614.1c self-replacement reading revolt (revolt.go) as the
// creature enters, so the counter is there before any state-based
// action or trigger looks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a4f99939-4b70-41ee-ad38-2dda56dbc9ce",
		Name:            "Narnam Renegade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Replacements:    []game.ReplacementEffect{SelfEntersWithCountersIfRevolt(game.CounterPlusOne, 1)},
	})
}
