package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Keeper of the Beasts — Creature — Human Wizard {G}{G}, 1/2 (Exodus, S82 OLD
// b737, #5626):
//
//	"{G}, {T}: Choose target opponent who controls more creatures than you
//	 do as you activate this ability. Create a 2/2 green Beast creature
//	 token."
//
// The count is read only as the ability is announced (CR 602.2b): the target
// clause carries MoreCreaturesAsActivated, which the activation gate and the
// legal set read and the CR 608.2b re-check does not, so the token still
// arrives if the counts change in response. The opponent has no other role.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86dcdffd-c14f-4e06-a74f-d2c727f55b1c",
		Name:         "Keeper of the Beasts",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: Choose target opponent who controls more creatures than you do as you activate this ability. Create a 2/2 green Beast creature token.",
			Purpose: game.Purpose{Tokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Targets: effects.WithMoreCreaturesAsActivated(effects.TargetPlayer("target opponent", effects.Opponent())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("2/2 green Beast"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
