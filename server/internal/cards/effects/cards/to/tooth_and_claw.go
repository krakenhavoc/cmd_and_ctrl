package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tooth and Claw — Enchantment {3}{R} (Eternal Masters, S82 OLD, #5542):
//
//	"Sacrifice two creatures: Create a 3/1 red Beast creature token named
//	 Carnivore."
//
// Whisper, Blood Liturgist's two-creature sacrifice cost, paid at announce
// (CR 602.2b), around the printed token; no limit on how often.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "587368eb-068c-44a3-ba8c-5ad0f59f880f",
		Name:         "Tooth and Claw",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice two creatures: Create a 3/1 red Beast creature token named Carnivore.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Cost:    effects.SacrificeN(2, "two creatures", effects.Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("3/1 red Beast named Carnivore"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
