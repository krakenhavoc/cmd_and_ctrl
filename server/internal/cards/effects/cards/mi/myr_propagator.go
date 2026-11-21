package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myr Propagator — Artifact Creature — Myr {3}, 1/1 (Scars of Mirrodin,
// #5577):
//
//	"{3}, {T}: Create a token that's a copy of this creature."
//
// Joo Dee's token copy of its own source (CR 707.2): the copiable values
// of the Propagator as it is when the ability resolves, or as it last
// existed on the battlefield if it has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d071e2fc-3be8-4168-afa9-ba34edad18b9",
		Name:         "Myr Propagator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Create a token that's a copy of this creature.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: ctx.Source(), N: 1}.Apply(ctx)
			})),
		}},
	})
}
