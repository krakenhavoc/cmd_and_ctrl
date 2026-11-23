package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Predatory Impetus — Enchantment — Aura {4}{G} (Ravnica: Clue Edition):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3, must be blocked if able, and is goaded.
//	 (It attacks each combat if able and attacks a player other than you if
//	 able.)"
//
// Coercive Impetus's pump and goad pair, plus the must-be-blocked requirement
// (CR 509.1c) on the enchanted creature, which belongs to the Aura and ends
// when it falls off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87f22ec6-6668-4e15-b38c-fa1ac0292c6b",
		Name:         "Predatory Impetus",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			predatoryImpetusMustBeBlocked(),
			effects.GoadAttached(),
		},
	})
}

func predatoryImpetusMustBeBlocked() game.StaticAbility {
	s := effects.BlockRequirementWhere(game.BlockRequirementMustBeBlocked, effects.AttachedToSource)
	s.Label = "Enchanted creature must be blocked if able."
	return s
}
