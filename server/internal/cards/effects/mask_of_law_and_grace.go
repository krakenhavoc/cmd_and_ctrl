package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mask of Law and Grace — Enchantment — Aura {W} (Urza's Destiny,
// #5621):
//
//	"Enchant creature
//	 Enchanted creature has protection from black and from red."
//
// A layer 6 grant to the enchanted creature (GrantToAttached), one
// protection quality per colour as on Paladin en-Vec (CR 702.16): no
// black or red creature can block it, black and red spells and
// abilities can't target it, and damage from black or red sources is
// prevented. The grant ends when the Aura leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "21474195-5d91-466d-8874-4818c900dea7",
		Name:         "Mask of Law and Grace",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			GrantToAttached(game.ProtectionFromColor("B"), game.ProtectionFromColor("R")),
		},
	})
}
