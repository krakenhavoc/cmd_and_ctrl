package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Acolyte of Xathrid — Creature — Human Cleric {B}, 0/1 (Magic 2010, #5585):
//
//	"{1}{B}, {T}: Target player loses 1 life."
//
// A tap ability that targets a player; the life is lost, not dealt as
// damage, by the player still legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62dca41b-58b0-45eb-ada7-568e80a2abec",
		Name:         "Acolyte of Xathrid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}: Target player loses 1 life.",
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B43TargetPlayerLoses(g, item, 1)
			},
		}},
	})
}
