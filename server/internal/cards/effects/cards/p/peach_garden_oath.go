package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peach Garden Oath — Sorcery {W} (Eighth Edition, #5607):
//
//	"You gain 2 life for each creature you control."
//
// The creatures its controller controls as it resolves (CR 608.2h), counted
// through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad66a99b-c7e7-4b2c-8525-299d3609d1df",
		Name:         "Peach Garden Oath",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OldB608PeachGardenOath,
	})
}
