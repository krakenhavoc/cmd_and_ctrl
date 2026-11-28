package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snake Basket — Artifact {4} (Visions, S82 OLD, #5631):
//
//	"{X}, Sacrifice this artifact: Create X 1/1 green Snake creature tokens.
//	 Activate only as a sorcery."
//
// X is announced with the activation and read from the ability as it
// resolves, so the tokens come even though the artifact is already gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12236231-7c6e-4edc-87dc-737115843185",
		Name:         "Snake Basket",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:        "{X}, Sacrifice this artifact: Create X 1/1 green Snake creature tokens. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{X}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.CreateToken{Template: effects.TokenCard("1/1 green Snake"), N: ctx.X()}.Apply(ctx)
			},
		}},
	})
}
