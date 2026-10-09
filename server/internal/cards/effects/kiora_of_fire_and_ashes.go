package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kiora of Fire and Ashes — Legendary Creature — Merfolk Noble
// {4}{R}{R}, 2/2:
//
//	"When Kiora enters, create a 5/5 red Dragon creature token with
//	 flying.
//	 {8}: Create a 5/5 red Dragon creature token with flying."
//
// No simplification.
func init() {
	dragon := Do(CreateToken{Template: TokenCard("5/5 red Dragon with flying"), N: 1})
	Register(Spec{
		OracleID:     "02b41d0e-82e7-4307-8cc4-8175ff78ea1f",
		Name:         "Kiora of Fire and Ashes",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kiora of Fire and Ashes — create a 5/5 red Dragon with flying", dragon),
		},
		Activated: []ActivatedAbility{{
			Label:  "{8}: Create a 5/5 red Dragon creature token with flying.",
			Cost:   ManaCost("{8}"),
			Effect: dragon,
		}},
	})
}
