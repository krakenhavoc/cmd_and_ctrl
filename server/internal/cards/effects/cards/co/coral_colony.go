package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coral Colony — Creature — Wall {1}{U}, 1/4 (Dominaria United):
//
//	"Defender
//	 {1}{U}, {T}: Target player mills X cards, where X is the number of
//	 creatures you control with defender."
//
// X is counted when the ability resolves, through the layers, and includes
// the Colony itself while it is on the battlefield (it has defender).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5e186931-992b-491c-bffc-6877a613d5c3",
		Name:            "Coral Colony",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Target player mills X cards, where X is the number of creatures you control with defender.",
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				victim, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return effects.MillCards{Player: victim, N: effects.CmdrB17DefendersControlled(g, ctx.Controller())}.Apply(ctx)
			},
		}},
	})
}
