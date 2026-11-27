package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Telethopter — Artifact Creature — Thopter {4}, 3/1 (Tempest, S82 OLD b754,
// #5628):
//
//	"Tap an untapped creature you control: This creature gains flying until
//	 end of turn."
//
// Bramblesnap's cost (TapAnUntapped: not the {T} symbol, so a creature that
// came under your control this turn may pay, and the Telethopter may tap
// itself) with a grant on the Telethopter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8202ff77-9afc-4e5b-895b-471bbfa67cce",
		Name:         "Telethopter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped creature you control: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"flying"}}},
			Cost:    effects.TapAnUntapped("an untapped creature you control", effects.Creature()),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Telethopter — flying until end of turn"),
		}},
	})
}
