package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pride of the Perfect — Enchantment {3}{B} (Commander Legends, #5507):
//
//	"Elves you control get +2/+0."
//
// A lord's anthem on an enchantment: TribalAnthem over the Elf creatures
// its controller controls, read through the layers, so a changeling
// counts and an Elf under another player does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83e2b251-bd60-4c5c-b9e2-32c22b28fb8b",
		Name:         "Pride of the Perfect",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Elf"}, YoursOnly: true}, 2, 0),
		},
	})
}
