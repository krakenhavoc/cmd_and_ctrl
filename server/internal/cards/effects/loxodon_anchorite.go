package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loxodon Anchorite — Creature — Elephant Cleric {2}{W}{W}, 2/3 (Fifth Dawn, #5604):
//
//	"{T}: Prevent the next 2 damage that would be dealt to any target
//	 this turn."
//
// Aven Redeemer's shield: a one-use prevention of 2 on the target that
// lasts until end of turn and is used up as it prevents (CR 615.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c70dded8-658b-4250-8057-54894350455c",
		Name:         "Loxodon Anchorite",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Prevent the next 2 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    TapCost(),
			Targets: TargetAny(),
			Effect:  preventNextDamageToTheTarget("Loxodon Anchorite", 2),
		}},
	})
}
