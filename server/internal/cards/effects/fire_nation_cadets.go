package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fire Nation Cadets — Creature — Human Soldier {R}, 1/2 (Avatar: The
// Last Airbender, #4459):
//
//	"This creature has firebending 2 as long as there's a Lesson card in
//	 your graveyard.
//	 {2}: This creature gets +1/+0 until end of turn."
//
// The firebending it has is an attack trigger, so the Lesson must be in
// your graveyard as it attacks (the ruling): a Lesson put there later
// gives the ability too late to trigger, and one that leaves after the
// trigger doesn't stop it (it is not an intervening "if").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ebfa3f5-f80a-4845-bba7-75c4ddb88682",
		Name:         "Fire Nation Cadets",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, AllOf(ThisAttacked, avatarB11LessonInYourGraveyard),
				"Fire Nation Cadets — firebending 2", Do(firebending(2))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Fire Nation Cadets — +1/+0 until end of turn"),
		}},
	})
}
