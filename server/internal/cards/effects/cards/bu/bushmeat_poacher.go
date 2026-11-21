package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bushmeat Poacher — Creature — Human Soldier {3}{B}, 2/4:
//
//	"{1}, {T}, Sacrifice another creature: You gain life equal to the
//	 sacrificed creature's toughness. Draw a card."
//
// The sacrifice is a cost, so the creature is gone before the ability
// resolves; its toughness is read from last-known information
// (departedCreatureToughness), the same read Consuming Vapors makes.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0287d541-3f73-4c8b-9a77-87f99898758d",
		Name:         "Bushmeat Poacher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice another creature: You gain life equal to the sacrificed creature's toughness. Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			// The draw waits for a prompt the life gain asks (a CR 616 order; CR 608.2c, #4155).
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					if fed, ok := effects.B17PermanentSacrificedToPay(ctx.Game, ctx.Item); ok {
						if t := effects.DepartedCreatureToughness(ctx.Game, fed); t > 0 {
							return effects.GainLife{Player: ctx.Controller(), Amount: t}.Apply(ctx)
						}
					}
					return nil
				}),
				effects.DrawCards{N: 1},
			),
		}},
	})
}
