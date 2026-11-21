package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viscerid Armor — Enchantment — Aura {1}{U} (Alliances, #5634):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1.
//	 {1}{U}: Return this Aura to its owner's hand."
//
// A layer 7c boost (CR 613.4c) to the enchanted creature, and an
// activated ability that bounces the Aura itself, to its owner's hand
// whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87b4cbec-3521-4bdb-bf01-64ddb526b655",
		Name:         "Viscerid Armor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpEnchantedCreature(1, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
