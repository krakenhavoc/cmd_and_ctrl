package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kumano's Blessing — Enchantment — Aura {2}{R}:
//
//	"Flash
//	 Enchant creature
//	 If a creature dealt damage by enchanted creature this turn would
//	 die, exile it instead."
//
// The replacement reads the turn's record of which objects the enchanted
// creature damaged (#1946, ExileIfDealtDamageByEnchantedWouldDie): the
// creature the Aura is attached to as the other creature would die, so
// damage it dealt before the Aura arrived counts (flash it in after
// combat damage), and damage dealt by a creature the Aura has since left
// does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "85d81888-8df7-4826-891e-a14fb8bf6549",
		Name:            "Kumano's Blessing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Replacements: []game.ReplacementEffect{
			ExileIfDealtDamageByEnchantedWouldDie("Kumano's Blessing: exile it instead of dying"),
		},
	})
}
