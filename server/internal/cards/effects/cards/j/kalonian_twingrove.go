package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kalonian Twingrove — Creature — Treefolk Warrior {5}{G}, */* (Magic 2015,
// #5556):
//
//	"Kalonian Twingrove's power and toughness are each equal to the number of
//	 Forests you control.
//	 When this creature enters, create a green Treefolk Warrior creature token
//	 with 'This token's power and toughness are each equal to the number of
//	 Forests you control.'"
//
// Dungrove Elder's characteristic-defining ability on the creature, and the
// same ability, read for its own controller, on the token it makes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbaf8e87-da5b-4120-afec-8fe2db55b835",
		Name:         "Kalonian Twingrove",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Static: []game.StaticAbility{
			effects.OldB356ForestsPT("Kalonian Twingrove's power and toughness are each equal to the number of Forests you control."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kalonian Twingrove — create a green Treefolk Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.OldB356TreefolkToken(), N: 1})),
		},
	})
}
