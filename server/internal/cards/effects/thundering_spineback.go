package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thundering Spineback — Creature — Dinosaur {5}{G}{G}, 5/5 (Lost
// Caverns of Ixalan Commander):
//
//	"Other Dinosaurs you control get +1/+1.
//	 {5}{G}: Create a 3/3 green Dinosaur creature token with trample."
//
// A lord over the controller's other Dinosaurs (CR 613.4c) and a plain
// mana-cost token maker; the 3/3 trample Dinosaur row is Regisaur Alpha's.
func init() {
	Register(Spec{
		OracleID:     "c9db31d7-00cf-4200-a11f-74cc2d9e843e",
		Name:         "Thundering Spineback",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Dinosaur"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}{G}: Create a 3/3 green Dinosaur creature token with trample.",
			Cost:    ManaCost("{5}{G}"),
			Purpose: game.Purpose{Tokens: 1},
			Effect:  Do(CreateToken{Template: TokenCard("3/3 green Dinosaur with trample"), N: 1}),
		}},
	})
}
