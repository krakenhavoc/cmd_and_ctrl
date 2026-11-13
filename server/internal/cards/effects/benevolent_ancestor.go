package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Benevolent Ancestor — Creature — Spirit {2}{W}, 0/4 (Ravnica: City of
// Guilds, #5600):
//
//	"Defender (This creature can't attack.)
//	 {T}: Prevent the next 1 damage that would be dealt to any target this
//	 turn."
//
// Kitsune Healer's first ability on a defender: a one-use shield of 1 on
// the target that lasts until end of turn (CR 615.7), charged as it
// prevents.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ee414d20-b18e-4bed-97fc-3c7535089ec1",
		Name:            "Benevolent Ancestor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    TapCost(),
			Targets: TargetAny(),
			Effect:  preventNextOneDamageToTheTarget("Benevolent Ancestor"),
		}},
	})
}
