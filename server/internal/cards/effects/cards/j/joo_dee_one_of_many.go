package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Joo Dee, One of Many — Creature — Human Advisor {1}{B}, 2/2 (Avatar:
// The Last Airbender, #4459):
//
//	"{B}, {T}: Surveil 1. Create a token that's a copy of this creature,
//	 then sacrifice an artifact or creature. Activate only as a
//	 sorcery."
//
// Three steps of one Do, each waiting for the one before (CR 608.2c):
// the surveil prompt is answered, then the token copy is made from Joo
// Dee's copiable values (CR 707.2; its last-known ones if it has left,
// CR 608.2h), then its controller sacrifices an artifact or creature,
// chosen as that step runs, so the new token and Joo Dee herself are
// among the choices (the first ruling). Nobody can act between the copy
// and the sacrifice. A copy of a token copy of Joo Dee is still Joo Dee.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4601ed7b-716e-4c89-b733-1eeabe8bf5a2",
		Name:         "Joo Dee, One of Many",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{B}, {T}: Surveil 1. Create a token that's a copy of this creature, then sacrifice an artifact or creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
			SorcerySpeed: true,
			Effect: effects.Do(
				effects.Surveil{N: 1},
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: ctx.Source(), N: 1}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.SacrificeChoice{
						Player:     ctx.Controller(),
						Candidates: effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Or(effects.Artifact(), effects.Creature())),
						Question:   "Joo Dee, One of Many — sacrifice an artifact or creature",
					}.Apply(ctx)
				}),
			),
		}},
	})
}
