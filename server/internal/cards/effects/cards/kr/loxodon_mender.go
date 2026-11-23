package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loxodon Mender — Creature — Elephant Cleric {5}{W}, 3/3 (Mirrodin, #5606):
//
//	"{W}, {T}: Regenerate target artifact."
//
// Welding Jar's regeneration of a target artifact, paid with {W} and
// a tap: the shield lands on the target as the ability resolves
// (CR 701.19a), and a target gone by then gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b260b5b4-048b-4a70-ab3a-b0dd1ceb7aa8",
		Name:         "Loxodon Mender",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Regenerate target artifact.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
