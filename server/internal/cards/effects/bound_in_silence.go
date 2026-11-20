package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bound in Silence — Kindred Enchantment — Rebel Aura {2}{W}
// (Future Sight, #5594):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block."
//
// Pacifism with the Rebel creature type (CR 308.4, a kindred card's
// subtypes are creature types), which rides the type line from Scryfall
// so a Rebel search or lord sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8cee34a-710d-4c09-b475-0e4754eccd66",
		Name:         "Bound in Silence",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock),
		},
	})
}
