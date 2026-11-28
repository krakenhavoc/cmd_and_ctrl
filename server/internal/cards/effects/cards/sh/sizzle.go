package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sizzle — Sorcery {2}{R} (Eighth Edition, S82 OLD batch 611, #5607):
//
//	"Sizzle deals 3 damage to each opponent."
//
// Boltwave's body at the same amount: one damage instance to every opponent
// still in the game, through the damage pipeline.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "938cc97f-3908-401e-94cd-32004aff1ef1",
		Name:         "Sizzle",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OldB611EachOpponentTakesThree,
	})
}
