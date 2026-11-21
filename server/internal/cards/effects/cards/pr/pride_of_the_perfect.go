package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "83e2b251-bd60-4c5c-b9e2-32c22b28fb8b",
		Name:         "Pride of the Perfect",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Elf"}, YoursOnly: true}, 2, 0),
		},
	})
}
