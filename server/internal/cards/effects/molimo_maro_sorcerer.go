package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molimo, Maro-Sorcerer — Legendary Creature — Elemental Sorcerer
// {4}{G}{G}{G}, */* (Commander Masters):
//
//	"Trample
//	 Molimo's power and toughness are each equal to the number of lands you
//	 control."
//
// Trample is printed data; the size is Ashaya's and Beanstalk Giant's
// characteristic-defining ability (CR 604.3, layer 7a), counted live.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b87a2d20-4e9f-4256-9d24-80c38515119e",
		Name:            "Molimo, Maro-Sorcerer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			landsYouControlPT("Molimo's power and toughness are each equal to the number of lands you control."),
		},
	})
}
