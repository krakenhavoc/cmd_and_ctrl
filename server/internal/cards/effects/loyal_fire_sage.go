package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loyal Fire Sage — Creature — Human Cleric Ally {2}{R}, 3/3 (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"Firebending 1 (Whenever this creature attacks, add {R}. This mana
//	 lasts until end of combat.)
//	 {5}: Create a 1/1 white Ally creature token."
//
// Firebending is the set's keyword trigger (firebending), its mana marked
// to last until end of combat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c052420-2d8b-4bbf-8197-b8a36e7208c9",
		Name:         "Loyal Fire Sage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Loyal Fire Sage — firebending 1", Do(firebending(1))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}: Create a 1/1 white Ally creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{5}"),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1}),
		}},
	})
}
