package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Leucrocota — Creature — Beast {3}{G}, 2/4 (Conspiracy: Take the
// Crown, S82 older sets batch 271, #5540):
//
//	"Vigilance
//	 {6}{G}: Monstrosity 3. (If this creature isn't monstrous, put three
//	 +1/+1 counters on it and it becomes monstrous.)"
//
// Vigilance is the engine's keyword (CR 702.20) and the activation is the
// shared Monstrosity (CR 701.37): once monstrous, a second activation does
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dcc5ca3f-5309-4833-bee4-f2508382ea06",
		Name:            "Ravenous Leucrocota",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{6}{G}"), 3)},
	})
}
