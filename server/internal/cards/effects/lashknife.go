package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lashknife — Enchantment — Aura {1}{W}:
//
//	"If you control a Plains, you may tap an untapped creature you
//	 control rather than pay this spell's mana cost.
//	 Enchant creature
//	 Enchanted creature has first strike."
//
// ADR 0135 §1 (#2030): the tap alternative cost (CR 118.9), offered only
// while you control a Plains, on Battle Mastery's shape with first
// strike. The tapped creature may be the one it enchants.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "98d790c0-985f-44a4-b247-0951beea7637",
		Name:         "Lashknife",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		AlternativeCosts: []game.AlternativeCost{
			TapInstead(1, "an untapped creature you control", ControlsA("Plains"), Creature()),
		},
		Static: []game.StaticAbility{
			GrantToAttached("first strike"),
		},
	})
}
