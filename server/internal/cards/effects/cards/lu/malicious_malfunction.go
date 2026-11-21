package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malicious Malfunction — Sorcery {1}{B}{B}:
//
//	"All creatures get -2/-2 until end of turn. If a creature would die
//	 this turn, exile it instead."
//
// Flaying Tendrils without devoid; see flaying_tendrils.go.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7ecaa1a-fbf7-436b-a1c9-d7810b0dc5dc",
		Name:         "Malicious Malfunction",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.AllCreaturesShrinkThenExileIfTheyDie(-2, false),
	})
}
