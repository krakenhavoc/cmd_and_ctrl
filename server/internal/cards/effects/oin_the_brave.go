package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Óin the Brave — Legendary Creature — Dwarf Warrior {1}{R}, 1/3:
//
//	"Storied (If you control three or more artifacts, legendaries,
//	 and/or Sagas, you have an enduring story for the rest of the game.)
//	 As long as you have an enduring story, Óin gets +1/+0 and has haste.
//	 {1}, {T}, Discard a card: Draw a card."
//
// Storied is the engine's (CR 702.195a). The haste is a layer-6 grant
// read by the layer pass, so an Óin that has been under your control
// since your turn began is unaffected and one that just arrived can tap
// for its own ability once the story is told (CR 302.6).
//
// No simplification. The Hobbit (#4475).
func init() {
	Register(Spec{
		OracleID:        "a25ba5cc-6ff8-4215-b70d-7e86e15e3c1a",
		Name:            "Óin the Brave",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordStoried},
		Static: []game.StaticAbility{
			SelfPumpWhileEnduringStory(1, 0),
			SelfKeywordWhileEnduringStory("haste"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Discard a card: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost(), DiscardACard()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
