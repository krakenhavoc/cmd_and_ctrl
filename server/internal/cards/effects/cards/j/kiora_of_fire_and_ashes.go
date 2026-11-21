package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kiora of Fire and Ashes — Legendary Creature — Merfolk Noble
// {4}{R}{R}, 2/2:
//
//	"When Kiora enters, create a 5/5 red Dragon creature token with
//	 flying.
//	 {8}: Create a 5/5 red Dragon creature token with flying."
//
// No simplification.
func init() {
	dragon := effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 red Dragon with flying"), N: 1})
	effects.Register(effects.Spec{
		OracleID:     "02b41d0e-82e7-4307-8cc4-8175ff78ea1f",
		Name:         "Kiora of Fire and Ashes",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kiora of Fire and Ashes — create a 5/5 red Dragon with flying", dragon),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}: Create a 5/5 red Dragon creature token with flying.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{8}"),
			Effect:  dragon,
		}},
	})
}
