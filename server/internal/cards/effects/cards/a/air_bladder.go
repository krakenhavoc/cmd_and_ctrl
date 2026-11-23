package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Air Bladder — Enchantment — Aura {U}
// (Nemesis, #5618):
//
//	"Enchant creature
//	 Enchanted creature has flying.
//	 Enchanted creature can block only creatures with flying."
//
// Angelic Gift's grant without the draw, plus the Perchers' block rule
// pointed at the enchanted creature
// (nemBlocksOnlyFlyers(OnAttached())). Both follow the Aura: once it
// leaves, the creature loses flying and may block anything again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25c8a3b0-c393-4fc6-a373-808cd8e9850c",
		Name:         "Air Bladder",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		BlockRules:   effects.NemBlocksOnlyFlyers(effects.OnAttached()),
	})
}
