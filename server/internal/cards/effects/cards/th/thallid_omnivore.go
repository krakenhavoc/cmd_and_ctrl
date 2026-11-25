package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thallid Omnivore — Creature — Fungus {3}{B}, 3/3 (Dominaria, S82 OLD batch
// 214, #5530):
//
//	"{1}, Sacrifice another creature: This creature gets +2/+2 until end of
//	 turn. If a Saproling was sacrificed this way, you gain 2 life."
//
// Boneyard Desecrator's shape: the sacrifice is the ordinary "another
// creature" cost, and "if a Saproling was sacrificed this way" reads the cost's
// record of the creature as it last existed on the battlefield (CR 608.2h), so
// a changeling counts and the life comes even when the Omnivore has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9bb1b0e-38ce-40f1-882b-85f6c95610ff",
		Name:         "Thallid Omnivore",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: This creature gets +2/+2 until end of turn. If a Saproling was sacrificed this way, you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump, LifeGain: 2},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.ThisGetsUntilEndOfTurn(2, 2, "Thallid Omnivore — +2/+2")(ctx.Game, ctx.Item)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					info, ok := ctx.SacrificedPermanent()
					if !ok || !(info.Characteristic.AllCreatureTypes || effects.HasFold(info.Characteristic.Subtypes, "Saproling")) {
						return nil
					}
					return effects.GainLife{Amount: 2}.Apply(ctx)
				}),
			),
		}},
	})
}
