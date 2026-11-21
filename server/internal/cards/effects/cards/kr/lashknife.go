package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "98d790c0-985f-44a4-b247-0951beea7637",
		Name:         "Lashknife",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		AlternativeCosts: []game.AlternativeCost{
			effects.TapInstead(1, "an untapped creature you control", effects.ControlsA("Plains"), effects.Creature()),
		},
		Static: []game.StaticAbility{
			effects.GrantToAttached("first strike"),
		},
	})
}
