package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spirit Mantle — Enchantment — Aura {1}{W}:
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has protection from creatures."
//
// Protection from a card type is in the closed protection grammar
// ("creatures"), so the granted token is a real protection: the
// enchanted creature can't be blocked by creatures, damaged by them,
// enchanted or equipped by creature permanents, or targeted by
// creature sources' abilities (CR 702.16).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81a51328-b995-4f3b-90bc-a20ae21dd254",
		Name:         "Spirit Mantle",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("protection from creatures"),
		},
	})
}
