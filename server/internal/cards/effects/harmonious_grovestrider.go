package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harmonious Grovestrider — Creature — Beast {3}{G}{G}, */* (Edge of
// Eternities, #4461):
//
//	"Ward {2}
//	 Harmonious Grovestrider's power and toughness are each equal to the
//	 number of lands you control."
//
// Ward is the shared Ward trigger (Aboleth Spawn's {2}), and the size is
// Ashaya's and Flora Colossus's characteristic-defining ability
// (landsYouControlPT, CR 604.3, layer 7a), counted live.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ebec97e-3cb1-41e2-a693-0febb5623016",
		Name:         "Harmonious Grovestrider",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			landsYouControlPT("Harmonious Grovestrider's power and toughness are each equal to the number of lands you control."),
		},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Harmonious Grovestrider — ward {2}"),
		},
	})
}
