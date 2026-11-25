package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Trenches — Enchantment {1}{R}{W} (Eternal Masters, S82 OLD, #5542):
//
//	"{2}, Sacrifice a land: Create two 1/1 red and white Goblin Soldier
//	 creature tokens."
//
// Ripchain Razorkin's "sacrifice a land" cost beside the mana, paid at
// announce; two tokens of the printed red-and-white Goblin Soldier.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b43f40e6-c0ad-4a12-b75d-f2ba12629bfe",
		Name:         "Goblin Trenches",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice a land: Create two 1/1 red and white Goblin Soldier creature tokens.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 2, CreatureTokens: 2},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeN(1, "a land", effects.Land())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 red and white Goblin Soldier"), N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
