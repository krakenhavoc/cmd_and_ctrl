package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unflinching Courage — Enchantment — Aura {1}{G}{W} (Foundations,
// #4465):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has trample and lifelink."
//
// A layer 7c pump and one layer 6 grant of both keywords to the
// enchanted creature, as Furious Strength and Felidar Umbra write them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e969a27-1609-4ab4-b0db-46b1af8066a9",
		Name:         "Unflinching Courage",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(2, 2), GrantToAttached("trample", "lifelink")},
	})
}
