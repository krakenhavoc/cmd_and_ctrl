package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "54dd35ee-6f89-460d-9370-72bc3fa6840a",
		Name:         "Part the Waterveil",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Awaken(6, "{6}{U}{U}{U}", nil),
		},
		OnResolve: effects.AwakenAfter(6, effects.TakeAnExtraTurnThenExileThisSpell),
	})
}
