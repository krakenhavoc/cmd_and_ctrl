package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greensleeves, Maro-Sorcerer — Legendary Creature — Elemental Sorcerer
// {3}{G}{G}, */* (Dominaria United Commander, S82):
//
//	"Protection from planeswalkers and from Wizards
//	 Greensleeves's power and toughness are each equal to the number of
//	 lands you control.
//	 Landfall — Whenever a land you control enters, create a 3/3 green
//	 Badger creature token."
//
// Two protection abilities (CR 702.16m), the characteristic-defining P/T
// over lands (layer 7a, the shared landsYouControlPT) and one landfall
// trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2ac2d483-2d2a-4912-8db4-04a5a2080831",
		Name:            "Greensleeves, Maro-Sorcerer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from planeswalkers", "protection from Wizards"},
		Static: []game.StaticAbility{
			landsYouControlPT("Greensleeves's power and toughness are each equal to the number of lands you control."),
		},
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Greensleeves, Maro-Sorcerer — create a 3/3 Badger (landfall)", game.LandfallPayoff{Tokens: 1}, Do(CreateToken{
				Template: TokenCard("3/3 green Badger"),
				N:        1,
			})),
		},
	})
}
