package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exoskeletal Armor — Enchantment — Aura {1}{G} (Judgment):
//
//	"Enchant creature
//	 Enchanted creature gets +X/+X, where X is the number of creature
//	 cards in all graveyards."
//
// Mortivore's count of creature cards in every graveyard, as a layer-7c
// modification of the enchanted creature rather than a characteristic-
// defining ability. Read on every layer pass, so it follows the graveyards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29358dcd-b26a-4720-9daa-7baa648e3540",
		Name:         "Exoskeletal Armor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{{
			Label:     "Enchanted creature gets +X/+X, where X is the number of creature cards in all graveyards.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := effects.B41CreatureCardsInAllGraveyards(g)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
