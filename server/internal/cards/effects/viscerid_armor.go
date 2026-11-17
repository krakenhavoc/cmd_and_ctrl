package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "87b4cbec-3521-4bdb-bf01-64ddb526b655",
		Name:         "Viscerid Armor",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{pumpEnchantedCreature(1, 1)},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
