package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b8cee34a-710d-4c09-b475-0e4754eccd66",
		Name:         "Bound in Silence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
		},
	})
}
