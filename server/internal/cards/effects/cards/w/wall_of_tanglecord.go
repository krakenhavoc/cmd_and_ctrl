package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Tanglecord — Artifact Creature — Wall {2}, 0/6 (Scars of
// Mirrodin, #5577):
//
//	"Defender
//	 {G}: This creature gains reach until end of turn."
//
// Defender is the engine's attack rule; the reach is a self grant until
// end of turn (somB1SelfGainsUntilEOT).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "423bebcc-8fa0-49c4-aa22-728bf51a535b",
		Name:            "Wall of Tanglecord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated:       []effects.ActivatedAbility{effects.SomB1SelfGainsUntilEOT("Wall of Tanglecord", "{G}", "reach")},
	})
}
