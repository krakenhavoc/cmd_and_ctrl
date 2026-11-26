package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ensoul Artifact — Enchantment — Aura {1}{U} (Magic 2015, #5556):
//
//	"Enchant artifact
//	 Enchanted artifact is a creature with base power and toughness 5/5 in
//	 addition to its other types."
//
// A layer 4 type addition and a layer 7b base P/T on the enchanted artifact
// (CR 613.1d, 613.4b), both read live off the attachment so they end with the
// Aura. The artifact is a creature in addition to its other types.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02628058-c9d8-4265-afd1-3fb697ec89f7",
		Name:         "Ensoul Artifact",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(effects.Artifact()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			{
				Label:     "Enchanted artifact is a creature in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasType(c.Types, "Creature") {
						c.Types = append(append([]string(nil), c.Types...), "Creature")
					}
				},
			},
			effects.SetAttachedBasePT(5, 5),
		},
	})
}
