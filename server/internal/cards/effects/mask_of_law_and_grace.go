package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mask of Law and Grace — Enchantment — Aura {W} (Urza's Destiny,
// #5621):
//
//	"Enchant creature
//	 Enchanted creature has protection from black and from red."
//
// Two protection grants on the enchanted creature, as the Swords print
// them (CR 702.16b). Protection from black and red doesn't touch the
// white Aura, so the Mask stays attached.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "21474195-5d91-466d-8874-4818c900dea7",
		Name:         "Mask of Law and Grace",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{GrantToAttached("protection from black", "protection from red")},
	})
}
