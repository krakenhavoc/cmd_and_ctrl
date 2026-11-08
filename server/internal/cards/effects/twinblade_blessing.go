package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twinblade Blessing — Enchantment — Aura {1}{W}{W} (Foundations,
// #4465):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature has double strike."
//
// Battle Mastery with flash, which Captain's Tutelage declares the same
// way: the Aura can be cast any time you could cast an instant, and the
// enchanted creature has double strike.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b3e1674f-dba1-4570-9990-e5dc643e237d",
		Name:            "Twinblade Blessing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{GrantToAttached("double strike")},
	})
}
