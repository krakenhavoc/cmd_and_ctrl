package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence Bottle — Artifact {2} (Tempest, S82 OLD batch 760, #5628):
//
//	"{3}, {T}: Put an elixir counter on this artifact.
//	 {T}, Remove all elixir counters from this artifact: You gain 2 life for
//	 each elixir counter removed this way."
//
// Jar of Eyeballs' "remove all counters" cost: the number removed is a fact
// of the announcement, read back as the ability resolves, so the life gained
// is right even if the Bottle has gone by then (CR 608.2h). With no counters
// the ability gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76f839b0-1d4c-4ad9-b44f-a8a13d6d109f",
		Name:         "Essence Bottle",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "{3}, {T}: Put an elixir counter on this artifact.",
				Cost:   effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				Effect: effects.PutACounterOnThis("elixir"),
			},
			{
				Label:   "{T}, Remove all elixir counters from this artifact: You gain 2 life for each elixir counter removed this way.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.TapCost(), effects.RemoveAllCountersFromThis("elixir")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := ctx.CountersRemoved()
					if n <= 0 {
						return nil
					}
					return effects.GainLife{Amount: 2 * n}.Apply(ctx)
				},
			},
		},
	})
}
