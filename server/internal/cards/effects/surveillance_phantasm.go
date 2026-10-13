package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surveillance Phantasm — Creature — Bird Illusion {1}{U}, 2/3:
//
//	"Defender, flying, vigilance
//	 As long as you've scried or surveilled this turn, this creature
//	 can attack as though it didn't have defender.
//	 {3}{U}: Surveil 1."
//
// "Can attack as though it didn't have defender" is a self-only layer 6
// static, Faithbound Judge's shape: the card has defender while its
// controller has not scried or surveilled this turn and does not while
// they have. It reads this turn's event log, and DependsOnEvents wakes
// the layer cache on a scry or a surveil, so the answer is live for one
// that happened before the Phantasm entered, and the turn's own advance
// clears it. The keyword is not in PrintedKeywords for that reason: the
// static is the only thing that grants it.
func init() {
	Register(Spec{
		OracleID:        "b97648b7-ae98-45c9-8f0f-a52d6faf2d2d",
		Name:            "Surveillance Phantasm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Static: []game.StaticAbility{{
			Layer:           game.Layer6Ability,
			Label:           "Has defender unless you've scried or surveilled this turn.",
			DependsOnEvents: []game.EventKind{game.EventScry, game.EventSurveil},
			AppliesTo:       selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if !rfCreatureBScriedOrSurveilledThisTurn(g, source.Controller) {
					appendKeywordsTo(c, []string{"defender"})
				}
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{3}{U}"),
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
