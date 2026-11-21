package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pestbrood Sloth — Creature — Plant Sloth {3}{G}, 4/4 (Secrets of
// Strixhaven, #4456):
//
//	"Reach
//	 When this creature dies, create two 1/1 black and green Pest creature
//	 tokens with "Whenever this token attacks, you gain 1 life.""
//
// The Pests are the Secrets of Strixhaven ones (PestAttackToken: life on
// attack, not the Strixhaven Pest's life on death). The tokens are the dead
// creature's controller's (CR 603.7c reads "you" off the ability's
// controller).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2970d09b-4bef-468e-8de5-f994c820ccc0",
		Name:            "Pestbrood Sloth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Pestbrood Sloth — create two 1/1 Pests that gain you life when they attack",
				effects.Do(effects.CreateToken{Template: effects.PestAttackToken(), N: 2})),
		},
	})
}
