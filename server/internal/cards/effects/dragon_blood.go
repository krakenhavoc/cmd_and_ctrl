package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon Blood — Artifact {3} (Urza's Saga):
//
//	"{3}, {T}: Put a +1/+1 counter on target creature."
//
// Tangle Tumbler's counter row on an artifact that is only that.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b751c1c6-195d-4023-9d0b-3c91d6b834d4",
		Name:         "Dragon Blood",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Put a +1/+1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  PlusOneCounterOnTheTarget,
		}},
	})
}
