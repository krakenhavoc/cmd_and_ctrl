package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undulating Witness — Creature — Serpent {4}{U}, 3/5:
//
//	"Flying
//	 {2}: This creature gets +1/-1 until end of turn.
//	 Basic landcycling {2}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "77b684b8-227b-40de-80b4-15d0e70914f4",
		Name:            "Undulating Witness",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: This creature gets +1/-1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{2}"),
				Effect:  effects.ThisCreatureUntilEOT("Undulating Witness — +1/-1", 1, -1),
			},
			effects.BasicLandcycling("{2}"),
		},
	})
}
