package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sanctum Weaver — Enchantment Creature — Dryad {1}{G}, 0/2:
//
//	"{T}: Add X mana of any one color, where X is the number of
//	 enchantments you control."
//
// Mona Lisa's shape (ProducedOneColor, one pick of N tokens) with a
// board count instead of a power read: N is the controller's own
// enchantments, Sanctum Weaver itself included (an enchantment
// creature is an enchantment). A count of zero adds nothing and still
// taps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acfe7ec0-0606-4d5e-b1fa-25f0c7aeec47",
		Name:         "Sanctum Weaver",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedOneColor(func(g *game.Game, controller, _ uuid.UUID) int {
				return b29EnchantmentsControlled(g, controller)
			}),
			Label: "Add X mana of any one color, where X is the number of enchantments you control",
		}},
	})
}
