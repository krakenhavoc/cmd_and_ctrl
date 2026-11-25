package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ichthyomorphosis — Enchantment — Aura {2}{U}:
//
//	"Enchant creature
//	 Enchanted creature loses all abilities and is a blue Fish with base
//	 power and toughness 0/1."
//
// Fish replaces the creature types only (CR 205.1a): the creature keeps its
// other card types, as Fog on the Barrow-Downs's Spirit does.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "816e13c6-79d8-431f-a04e-15c423baf7b3",
		Name:         "Ichthyomorphosis",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.SetAttachedCreatureTypes("Fish"),
			effects.SetAttachedColors("U"),
			effects.LoseAllAbilities(),
			effects.SetAttachedBasePT(0, 1),
		},
	})
}
