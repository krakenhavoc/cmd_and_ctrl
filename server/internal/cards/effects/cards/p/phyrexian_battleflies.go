package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Battleflies — Creature — Phyrexian Insect {B}, 0/1 (Invasion, S82
// OLD #5616):
//
//	"Flying
//	 {B}: This creature gets +1/+0 until end of turn. Activate no more than
//	 twice each turn."
//
// Roterothopter's shape: the limit is AtMostNActivationsPerTurn, the
// per-object count of announcements this turn (CR 602.1b), so a third
// activation is never offered, whether or not the first two have resolved.
// The pump is a self boost until end of turn.
//
// No simplification.
func init() {
	const label = "{B}: This creature gets +1/+0 until end of turn. Activate no more than twice each turn."
	effects.Register(effects.Spec{
		OracleID:        "58ed1294-ba7d-43a1-b988-8d78d9fa0e3e",
		Name:            "Phyrexian Battleflies",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{B}"),
			Condition: effects.AtMostNActivationsPerTurn(label, 2),
			Effect:    effects.ThisGetsUntilEndOfTurn(1, 0, "Phyrexian Battleflies — +1/+0 until end of turn"),
		}},
	})
}
