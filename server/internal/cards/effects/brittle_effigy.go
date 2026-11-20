package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brittle Effigy — Artifact {1} (Magic 2011, #5579):
//
//	"{4}, {T}, Exile this artifact: Exile target creature."
//
// Gerrard's Hourglass Pendant's cost shape (mana, tap and exiling the
// artifact itself, all paid as the ability is activated, CR 602.2b) with
// Catapult Master's effect: the target creature is exiled as the ability
// resolves, and a target gone by then is not (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34909cf0-780d-4bc7-9926-9f7625b8f200",
		Name:         "Brittle Effigy",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}, Exile this artifact: Exile target creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{4}"), TapCost(), ExileThis()),
			Targets: TargetCreature("target creature"),
			Effect:  exileFirstLegalCardTarget,
		}},
	})
}
