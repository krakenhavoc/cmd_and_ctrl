package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Smokebraider — Creature — Elemental Shaman {1}{R}, 1/1 (Lorwyn Eclipsed
// Commander, #5368):
//
//	"{T}: Add two mana in any combination of colors. Spend this mana only to
//	 cast Elemental spells or activate abilities of Elementals."
//
// A mana ability (no stack) of Axebane Guardian's any-combination shape with
// a fixed count of two, and Eclipsed Realms' spend restriction fixed on
// Elemental. The {T} is gated by summoning sickness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e970f2df-90cb-45b8-8ae5-71b762908925",
		Name:         "Smokebraider",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedAnyCombinationOfColors(func(_ *game.Game, _, _ uuid.UUID) int { return 2 }),
			Restrictions:   effects.Cmdr2S12ElementalManaRestrictions(),
			Label:          "Add two mana in any combination of colors (Elemental spells and Elementals' abilities only)",
		}},
	})
}
