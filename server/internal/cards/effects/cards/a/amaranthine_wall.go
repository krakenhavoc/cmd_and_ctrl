package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amaranthine Wall — Artifact Creature — Wall {4}, 0/6 (Dominaria, S82 OLD,
// #5530):
//
//	"Defender
//	 {2}: This creature gains indestructible until end of turn."
//
// A mana-only ability with no tap, so it can be activated any number of times
// and in response to damage or destruction; the gain goes to this permanent.
//
// No simplification.
func init() {
	const label = "Amaranthine Wall — gains indestructible until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "3c9fcda6-3ebc-4f9d-9f73-2bb4252d9420",
		Name:            "Amaranthine Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{2}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("indestructible", label),
		}},
	})
}
