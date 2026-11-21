package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vindictive Flamestoker — Creature — Phyrexian Wizard {R}, 1/2 (Phyrexia: All Will Be One):
//
//	"Whenever you cast a noncreature spell, put an oil counter on this creature.
//	 {6}{R}, Sacrifice this creature: Discard your hand, then draw four cards.
//	 This ability costs {1} less to activate for each oil counter on this creature."
//
// The reduction is the ability's own cost clause, generic mana only (CR 601.2f),
// read from the Flamestoker's live counters as it is priced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ed9a887-9942-4482-942e-b3b05feb6342",
		Name:         "Vindictive Flamestoker",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OneB09OilOnNoncreatureCast("Vindictive Flamestoker")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}{R}, Sacrifice this creature: Discard your hand, then draw four cards. This ability costs {1} less to activate for each oil counter on this creature.",
			Purpose: game.Purpose{Draws: 4, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{6}{R}"), effects.SacrificeThis()),
			CostModifiers: []game.CostModifier{effects.CostsLessEach(effects.CountersOnSource(effects.S82OneB09Oil),
				"This ability costs {1} less to activate for each oil counter on this creature.")},
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					_, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
					return err
				}),
				effects.DrawCards{N: 4},
			),
		}},
	})
}
