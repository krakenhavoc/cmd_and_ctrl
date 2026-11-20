package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Night Market Aeronaut — Creature — Aetherborn Warrior {3}{B}, 2/2 (Aether Revolt,
// #5536):
//
//	"Flying
//	 Revolt — This creature enters with a +1/+1 counter on it if a
//	 permanent left the battlefield under your control this turn."
//
// Narnam Renegade's revolt self-replacement (revolt.go, CR 614.1c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e08157e6-a5a3-4b1e-80e0-32702288d8ed",
		Name:            "Night Market Aeronaut",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{SelfEntersWithCountersIfRevolt(game.CounterPlusOne, 1)},
	})
}
