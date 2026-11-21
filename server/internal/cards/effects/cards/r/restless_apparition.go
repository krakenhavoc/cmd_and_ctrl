package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restless Apparition — Creature — Spirit {W/B}{W/B}{W/B}, 2/2:
//
//	"{W/B}{W/B}{W/B}: This creature gets +3/+3 until end of turn.
//	 Persist"
//
// Persist is PrintedKeywords (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fbd6dda7-ccee-4640-805d-78a9d60f45ef",
		Name:            "Restless Apparition",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W/B}{W/B}{W/B}: This creature gets +3/+3 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 3, Toughness: 3}},
			Cost:    effects.ManaCost("{W/B}{W/B}{W/B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(3, 3, "Restless Apparition — +3/+3 until end of turn"),
		}},
	})
}
