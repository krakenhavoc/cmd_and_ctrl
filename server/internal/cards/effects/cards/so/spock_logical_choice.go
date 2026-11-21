package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spock, Logical Choice — Legendary Creature — Vulcan Human Scientist
// {3}{U}{U}, 3/5 (Star Trek Commander, #4974):
//
//	"Vigilance
//	 {T}: Draw a card."
//
// Vigilance is the printed keyword. The tap ability has the {T} symbol,
// so it can't be activated until Spock has been under your control
// continuously since your most recent turn began (CR 302.6); after that,
// any time you have priority.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a877820f-8999-4dfe-a50d-09f63c885ce3",
		Name:            "Spock, Logical Choice",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
