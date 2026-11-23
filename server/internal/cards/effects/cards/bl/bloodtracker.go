package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodtracker — Creature — Vampire Wizard {3}{B}, 2/2 (Commander,
// #5368):
//
//	"Flying
//	 {B}, Pay 2 life: Put a +1/+1 counter on this creature.
//	 When this creature leaves the battlefield, draw a card for each
//	 +1/+1 counter on it."
//
// The leave trigger is Toothy, Imaginary Friend's: it counts the counters
// Bloodtracker had as it left (last-known information, CR 608.2h). The
// life is a cost (CR 119.4): you can't pay more than you have.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fcf9ced2-867d-4bc5-b7b6-6153ed2cac97",
		Name:            "Bloodtracker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Pay 2 life: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.PayLife(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Bloodtracker — draw a card for each +1/+1 counter on it",
				effects.Cmdr2B09DrawForEachCounterOnIt),
		},
	})
}
