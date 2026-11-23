package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whisk Away — Instant {2}{U} (Fate Reforged,
// #5552):
//
//	"Put target attacking or blocking creature on top of its owner's
//	 library."
//
// Time Ebb's tuck with the clause narrowed to an attacking or blocking
// creature, judged when it is targeted and again as the spell resolves (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac141fe0-328a-4d9b-8048-b0c47f399587",
		Name:         "Whisk Away",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve:    effects.PutTargetOnTopOfOwnersLibrary,
	})
}
