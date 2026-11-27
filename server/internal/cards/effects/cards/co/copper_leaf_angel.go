package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Copper-Leaf Angel — Artifact Creature — Angel {5}, 2/2 (Prophecy, S82 OLD
// batch 675, #5617):
//
//	"Flying
//	 {T}, Sacrifice X lands: Put X +1/+1 counters on this creature."
//
// Flying is a printed keyword. The cost is {T} and a sacrifice of exactly X
// lands, X announced with the activation (CR 602.2b; Grim Hireling's
// SacrificeX over lands); the effect reads that X back and puts that many
// +1/+1 counters on the Angel through the counter pipeline. X = 0 sacrifices
// nothing and adds nothing. The {T} is a creature's, so it waits out summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "efab5a30-74ee-4959-b7cd-2a5d25d18327",
		Name:            "Copper-Leaf Angel",
		Completeness:    effects.CompletenessFull,
		XMatters:        true,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice X lands: Put X +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeX("X lands", effects.Land())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if x := ctx.X(); x > 0 {
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
