package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strength of Lunacy — Enchantment — Aura {1}{B} (Torment, #5612):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+1 and has protection from white.
//	 Madness {B} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Strength of Isolation's black twin: a layer 7c pump (CR 613.4c) and a
// granted protection from white (CR 702.16a), which never knocks off this
// black Aura (CR 702.16c). Madness is the engine's (CR 702.35a, #657): a
// discarded Strength of Lunacy is exiled and offered for {B}, cast at
// instant speed as an ordinary Aura spell that targets the creature it
// will enchant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02a95ea7-cfb9-497c-a004-3e783856d79c",
		Name:         "Strength of Lunacy",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Madness:      "{B}",
		Static: []game.StaticAbility{
			PumpAttached(2, 1),
			GrantToAttached("protection from white"),
		},
	})
}
