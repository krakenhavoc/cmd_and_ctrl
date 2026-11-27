package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pit Imp — Creature — Imp {B}, 0/1 (Tempest, S82 OLD b755, #5628):
//
//	"Flying
//	 {B}: This creature gets +1/+0 until end of turn. Activate no more than
//	 twice each turn."
//
// Roterothopter's limit (AtMostNActivationsPerTurn: announcements this turn,
// CR 602.1b) on a +1/+0 self pump.
//
// No simplification.
func init() {
	const label = "{B}: This creature gets +1/+0 until end of turn. Activate no more than twice each turn."
	effects.Register(effects.Spec{
		OracleID:        "3decbc45-3e79-455f-9d62-878ee31b458d",
		Name:            "Pit Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{B}"),
			Condition: effects.AtMostNActivationsPerTurn(label, 2),
			Effect:    effects.ThisGetsUntilEndOfTurn(1, 0, "Pit Imp — +1/+0 until end of turn"),
		}},
	})
}
