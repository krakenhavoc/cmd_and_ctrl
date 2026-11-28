package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Master of Arms — Creature — Human Soldier {2}{W}, 2/2 (Weatherlight):
//
//	"First strike
//	 {1}{W}: Tap target creature blocking this creature."
//
// The target must be blocking this creature when the ability is put on the
// stack and again as it resolves (CR 608.2b); a blocker removed from combat in
// response is an illegal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "37e90cab-d904-4387-9d18-62f125c38c46",
		Name:            "Master of Arms",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Tap target creature blocking this creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.BlockingSource(effects.TargetCreature("target creature blocking this creature")),
			Effect:  effects.ThousandMoonsCrackshotTap,
		}},
	})
}
