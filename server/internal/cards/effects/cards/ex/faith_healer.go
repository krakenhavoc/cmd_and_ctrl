package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faith Healer — Creature — Human Cleric {1}{W}, 1/1 (Urza's Saga, S82 OLD
// batch 732, #5624):
//
//	"Sacrifice an enchantment: You gain life equal to the sacrificed
//	 enchantment's mana value."
//
// The enchantment is sacrificed as a cost (CR 601.2h), any enchantment its
// controller controls, the Healer included if it is one. The life is the
// mana value it last had on the battlefield (CR 608.2h; a token that is no
// copy has none and gains 0), through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c657e110-c204-48df-a396-9822ac48798d",
		Name:         "Faith Healer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an enchantment: You gain life equal to the sacrificed enchantment's mana value.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.SacrificeN(1, "an enchantment", effects.Enchantment()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.GainLife{Amount: effects.SacrificedManaValue(ctx)}.Apply(ctx)
			},
		}},
	})
}
