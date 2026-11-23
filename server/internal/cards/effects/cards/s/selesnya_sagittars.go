package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selesnya Sagittars — Creature — Elf Archer {3}{G}{W}, 2/5 (Ravnica: City
// of Guilds, #5600):
//
//	"Reach (This creature can block creatures with flying.)
//	 This creature can block an additional creature each combat."
//
// Reach is printed data; the extra block is Foriysian Brigade's
// CanBlockAdditional (CR 509.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "027484b8-0df5-48b6-b69b-bb4e39c5ba81",
		Name:            "Selesnya Sagittars",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static:          []game.StaticAbility{effects.CanBlockAdditional(effects.SelfOnly, 1)},
	})
}
