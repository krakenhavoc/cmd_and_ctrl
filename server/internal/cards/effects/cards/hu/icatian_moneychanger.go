package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icatian Moneychanger — Creature — Human {W}, 0/2 (Fallen Empires, S82 OLD
// b853, #5637):
//
//	"This creature enters with three credit counters on it.
//	 When this creature enters, it deals 3 damage to you.
//	 At the beginning of your upkeep, put a credit counter on this creature.
//	 Sacrifice this creature: You gain 1 life for each credit counter on this
//	 creature. Activate only during your upkeep."
//
// The counters are an entry replacement; the life is counted from the creature
// as it last existed, since sacrificing it is the cost (CR 608.2h). The 3 damage
// is dealt by the Moneychanger to its controller.
//
// No simplification.
func init() {
	const credit = "credit"
	effects.Register(effects.Spec{
		OracleID:     "9b81aa32-e0f7-493f-92b8-85c2c46aef91",
		Name:         "Icatian Moneychanger",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(credit, 3, "Icatian Moneychanger: enters with three credit counters")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Icatian Moneychanger — it deals 3 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.DealDamage{Source: ctx.Source(), Target: ctx.Controller(), Amount: 3}.Apply(ctx)
				}),
			effects.AtYourUpkeep("Icatian Moneychanger — put a credit counter on this creature", effects.PutACounterOnThis(credit)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "Sacrifice this creature: You gain 1 life for each credit counter on this creature. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.SacrificeThis(),
			Condition: effects.DuringYourUpkeep(),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Amount: effects.CountersOnThisAsItResolves(credit)(ctx)}.Apply(ctx)
			})),
		}},
	})
}
