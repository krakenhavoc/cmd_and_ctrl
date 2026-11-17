package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "25c8a3b0-c393-4fc6-a373-808cd8e9850c",
		Name:         "Air Bladder",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{GrantToAttached("flying")},
		BlockRules:   nemBlocksOnlyFlyers(OnAttached()),
	})
}
