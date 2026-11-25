package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Etherium-Horn Sorcerer — Artifact Creature — Minotaur Wizard Sorcerer
// {4}{U}{R}, 3/6 (Commander 2017, S82 older sets):
//
//	"{1}{U}{R}: Return this creature to its owner's hand.
//	 Cascade"
//
// Lantern Spirit's return-self ability, and Enigma Sphinx's cascade.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01afb7a1-6427-4bb6-935a-613e9bda9de3",
		Name:            "Etherium-Horn Sorcerer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"cascade"},
		Triggered:       []game.TriggeredAbility{effects.Cascade()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}{R}: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{U}{R}"),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
