package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thundering Spineback — Creature — Dinosaur {5}{G}{G}, 5/5 (Lost
// Caverns of Ixalan Commander):
//
//	"Other Dinosaurs you control get +1/+1.
//	 {5}{G}: Create a 3/3 green Dinosaur creature token with trample."
//
// A lord over the controller's other Dinosaurs (CR 613.4c) and a plain
// mana-cost token maker; the 3/3 trample Dinosaur row is Regisaur Alpha's.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c9db31d7-00cf-4200-a11f-74cc2d9e843e",
		Name:         "Thundering Spineback",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Dinosaur"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}: Create a 3/3 green Dinosaur creature token with trample.",
			Cost:    effects.ManaCost("{5}{G}"),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Dinosaur with trample"), N: 1}),
		}},
	})
}
