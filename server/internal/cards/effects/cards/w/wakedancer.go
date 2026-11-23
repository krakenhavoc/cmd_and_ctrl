package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wakedancer — Creature — Human Shaman {2}{B}, 2/2 (Foundations Jumpstart,
// S82, #5929):
//
//	"Morbid — When this creature enters, if a creature died this turn,
//	 create a 2/2 black Zombie creature token."
//
// Slumbering Cerberus's morbid intervening "if" (CR 603.4,
// condACreatureDiedThisTurn: any creature, anyone's, tokens included),
// checked as it enters and again as the trigger resolves, around an
// ordinary token trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e850743e-e4c4-4a09-b797-d5f688b51451",
		Name:         "Wakedancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.IfVentureCond(effects.CondACreatureDiedThisTurn,
				effects.WhenThisEnters("Wakedancer — morbid: create a 2/2 black Zombie creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}))),
		},
	})
}
