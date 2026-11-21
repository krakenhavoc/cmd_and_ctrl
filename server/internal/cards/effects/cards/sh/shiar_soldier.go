package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shi'ar Soldier — Creature — Shi'ar Soldier {1}{U}, 1/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flying
//	 {U}, {T}: Return another target permanent you control to its owner's
//	 hand. Activate only during your turn."
//
// Kelpie Guide's "another target permanent you control" on a {T} ability,
// the turn gate is DuringYourTurn, and the bounce is the simultaneous one
// Repulsor Bots uses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f20ec8e2-d6e4-4707-9b86-107dbc09ac63",
		Name:            "Shi'ar Soldier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{U}, {T}: Return another target permanent you control to its owner's hand. Activate only during your turn.",
			Cost:      effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets:   effects.Another(effects.TargetPermanent("another target permanent you control", effects.YouControl())),
			Condition: effects.DuringYourTurn(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B35BounceChosenPermanents(effects.NewContext(g, item))
			},
		}},
	})
}
