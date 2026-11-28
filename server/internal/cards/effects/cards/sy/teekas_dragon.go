package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teeka's Dragon — Artifact Creature — Dragon {9}, 5/5 (Mirage, #5633):
//
//	"Flying; trample; rampage 4 (Whenever this creature becomes blocked, it gets
//	 +4/+4 until end of turn for each creature blocking it beyond the first.)"
//
// Rampage is the engine's own keyword (CR 702.23).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6fb6e37e-d774-4a50-8ea3-b955535c2065",
		Name:            "Teeka's Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "rampage 4"},
	})
}
