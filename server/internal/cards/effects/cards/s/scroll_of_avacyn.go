package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scroll of Avacyn — Artifact {1} (Jumpstart, #5510):
//
//	"{1}, Sacrifice this artifact: Draw a card. If you control an Angel, you
//	 gain 5 life."
//
// The Angel is read as the ability resolves, after the sacrifice, through
// the layered view, so a changeling counts. Scroll of Avacyn is no Angel.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f728c4f1-178f-421b-a713-21f85892f051",
		Name:         "Scroll of Avacyn",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this artifact: Draw a card. If you control an Angel, you gain 5 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Effect: effects.Do(effects.DrawCards{N: 1}, effects.If{
				Cond: effects.CondBoard{Name: "you control an Angel", Test: func(ctx *effects.Context) bool {
					return effects.ControlsA(ctx.TextSubtype("Angel"))(ctx.Game, ctx.Controller())
				}},
				Then: []effects.Applier{effects.GainLife{Amount: 5}},
			}),
		}},
	})
}
