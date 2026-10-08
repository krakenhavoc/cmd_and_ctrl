package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Part the Waterveil — Sorcery {4}{U}{U}:
//
//	"Take an extra turn after this one. Exile Part the Waterveil.
//	 Awaken 6—{6}{U}{U}{U}"
//
// ADR 0135 §3 (#2411): the extra turn, the exile (part of the resolution,
// as Temporal Mastery's is), then the awaken land (CR 702.113a), which
// keeps its haste on the extra turn.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "54dd35ee-6f89-460d-9370-72bc3fa6840a",
		Name:         "Part the Waterveil",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Awaken(6, "{6}{U}{U}{U}", nil),
		},
		OnResolve: AwakenAfter(6, takeAnExtraTurnThenExileThisSpell),
	})
}
