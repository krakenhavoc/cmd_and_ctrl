package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Night Scythe — Artifact — Vehicle {3}, 3/1 (Warhammer 40,000 Commander,
// S82 GAP batch 186, #5948):
//
//	"Flying
//	 Invasion Beams — When this Vehicle enters, create a 2/2 black Necron
//	 Warrior artifact creature token.
//	 Crew 2"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dea9741a-3360-4a17-aeea-a854e2174d25",
		Name:            "Night Scythe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Night Scythe"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Night Scythe — create a 2/2 black Necron Warrior artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Necron Warrior artifact"), N: 1})),
		},
	})
}
