package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "a877820f-8999-4dfe-a50d-09f63c885ce3",
		Name:            "Spock, Logical Choice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    TapCost(),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
