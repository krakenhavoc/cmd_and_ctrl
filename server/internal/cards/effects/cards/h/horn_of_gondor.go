package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horn of Gondor — Legendary Artifact {3}:
//
//	"When Horn of Gondor enters, create a 1/1 white Human Soldier
//	 creature token.
//	 {3}, {T}: Create X 1/1 white Human Soldier creature tokens, where
//	 X is the number of Humans you control."
//
// X is counted as the ability resolves, not when it is activated, so
// a Human that dies in response is not counted and one that arrives
// is. The Soldier tokens are Humans themselves, so the Horn feeds
// on its own output.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88777fc0-286d-47fc-8e94-1e289b4964c4",
		Name:         "Horn of Gondor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Horn of Gondor — create a 1/1 white Human Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Create X 1/1 white Human Soldier creature tokens, where X is the number of Humans you control",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := effects.B43CreaturesOfSubtypeControlled(g, item.Controller, g.TextReaderForEffect(item).TextSubtype("Human"))
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: n}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
