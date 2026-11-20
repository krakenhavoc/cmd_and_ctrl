package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greenwheel Liberator — Creature — Elf Warrior {1}{G}, 2/1 (Aether Revolt,
// #5536):
//
//	"Revolt — This creature enters with two +1/+1 counters on it if a
//	 permanent left the battlefield under your control this turn."
//
// Narnam Renegade's revolt self-replacement with two counters
// (revolt.go, CR 614.1c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "121a8717-4c5b-45a8-875a-2b802c1fe6f8",
		Name:         "Greenwheel Liberator",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersWithCountersIfRevolt(game.CounterPlusOne, 2)},
	})
}
