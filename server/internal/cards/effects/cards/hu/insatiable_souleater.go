package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Insatiable Souleater — Artifact Creature — Phyrexian Beast {4}, 5/1 (New Phyrexia, #5575):
//
//	"{G/P}: This creature gains trample until end of turn. ({G/P} can be paid with either {G} or 2 life.)"
//
// A self grant until end of turn (somB1SelfGainsUntilEOT, Vector Asp's
// row); the Phyrexian symbol is paid with {G} or 2 life (CR 107.4f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f81c0ddc-cd3c-42c0-a7d9-0057b69a5df1",
		Name:         "Insatiable Souleater",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SomB1SelfGainsUntilEOT("Insatiable Souleater", "{G/P}", "trample")},
	})
}
