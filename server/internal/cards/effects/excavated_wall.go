package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Excavated Wall — Artifact Creature — Wall {1}, 0/4 (Strixhaven: School
// of Mages, #5502):
//
//	"Defender
//	 {1}, {T}: Mill a card. (Put the top card of your library into your
//	 graveyard.)"
//
// Defender is the printed keyword; the row is Waterlogged Hulk's mill with
// {1} added to the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "16e48b95-f030-400c-a2b7-c3f3111cc165",
		Name:            "Excavated Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Mill a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect:  Do(MillCards{N: 1}),
		}},
	})
}
