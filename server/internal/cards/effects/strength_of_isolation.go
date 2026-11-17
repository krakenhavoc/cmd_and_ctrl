package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strength of Isolation — Enchantment — Aura {1}{W} (Torment, #5612):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+2 and has protection from black.
//	 Madness {W} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Spirit Mantle's two statics: a layer 7c pump (CR 613.4c) and a granted
// protection from a colour, in the closed protection grammar (CR
// 702.16a). The Aura is white, so the protection never knocks it off
// (CR 702.16c). Madness is the engine's (CR 702.35a, #657), grown from
// the one string: a discarded Strength of Isolation is exiled and offered
// for {W}, cast at instant speed as an ordinary Aura spell that targets
// the creature it will enchant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87aad9fb-e234-4171-a4ae-c686f613d37e",
		Name:         "Strength of Isolation",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Madness:      "{W}",
		Static: []game.StaticAbility{
			PumpAttached(1, 2),
			GrantToAttached("protection from black"),
		},
	})
}
