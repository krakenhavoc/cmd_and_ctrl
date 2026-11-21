package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meldweb Strider — Artifact — Vehicle {4}{U}, 5/5 (Phyrexia: All Will Be
// One):
//
//	"Vigilance
//	 This Vehicle enters with an oil counter on it.
//	 Remove an oil counter from this Vehicle: It becomes an artifact
//	 creature until end of turn.
//	 Crew 3"
//
// The oil ability is a second way to animate it, at any time it has a
// counter, and needs no crew (CR 702.122). Like a crewed Vehicle it keeps
// its printed 5/5 and the Vehicle subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b66dd751-fef7-4310-9c09-8d1b1fc78fef",
		Name:            "Meldweb Strider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Replacements:    []game.ReplacementEffect{effects.B10EntersWithCounters(effects.OilCounter, 1, "Meldweb Strider: enters with an oil counter on it")},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Remove an oil counter from this Vehicle: It becomes an artifact creature until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    effects.RemoveCountersFromThis(effects.OilCounter, 1),
				Effect:  effects.CrewEffect("Meldweb Strider"),
			},
			{
				Label:   "Crew 3",
				Cost:    effects.CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Meldweb Strider"),
			},
		},
	})
}
