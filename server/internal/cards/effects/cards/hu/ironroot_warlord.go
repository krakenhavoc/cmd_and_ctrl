package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ironroot Warlord — Creature — Treefolk Soldier {1}{G}{W}, */5 (Magic
// 2014; Jumpstart):
//
//	"Ironroot Warlord's power is equal to the number of creatures you
//	 control.
//	 {3}{G}{W}: Create a 1/1 white Soldier creature token."
//
// Adeline's characteristic-defining ability (Layer 7a, the Warlord
// itself counted), and an instant-speed Soldier that grows it by one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "abc49dc3-03b3-48a4-afe1-5fc435b65d0f",
		Name:         "Ironroot Warlord",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.JmpB5PowerIsCreaturesYouControl("Ironroot Warlord's power is equal to the number of creatures you control."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}{W}: Create a 1/1 white Soldier creature token.",
			Cost:    effects.ManaCost("{3}{G}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Soldier"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
