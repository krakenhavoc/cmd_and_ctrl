package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dolmen Gate — Artifact {2}:
//
//	"Prevent all combat damage that would be dealt to attacking creatures
//	 you control."
//
// Goldbug's and Losheel's standing CR 615 prevention with the kind widened
// to every creature: combat damage only, attacking creatures only, this
// permanent's controller's only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d53d000f-c92f-4d70-ae60-8be3791da8a2",
		Name:         "Dolmen Gate",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			preventCombatDamageToYourAttackers(
				"Dolmen Gate: prevent all combat damage that would be dealt to attacking creatures you control",
				func(game.Card) bool { return true }),
		},
	})
}
