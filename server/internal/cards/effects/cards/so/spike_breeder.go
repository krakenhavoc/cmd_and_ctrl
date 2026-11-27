package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spike Breeder — Creature — Spike {3}{G}, 0/0 (Stronghold, S82 OLD batch 748,
// #5627):
//
//	"This creature enters with three +1/+1 counters on it.
//	 {2}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on
//	 target creature.
//	 {2}, Remove a +1/+1 counter from this creature: Create a 1/1 green Spike
//	 creature token."
//
// The Spike cycle's shared entry replacement and counter move
// (spike_counters.go) and a token row paid with the same cost: the counter is
// removed as the ability is announced (CR 118.3), so the Breeder can pay its
// last one and leave a 0/0 that dies with the ability still on the stack.
//
// No simplification.
func init() {
	const name = "Spike Breeder"
	effects.Register(effects.Spec{
		OracleID:     "b7b738e1-298d-489d-9939-efe1c983f0ae",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SpikeEntersWith(name, "three", 3)},
		Activated: []effects.ActivatedAbility{
			effects.SpikeMovesACounter(),
			{
				Label:   "{2}, Remove a +1/+1 counter from this creature: Create a 1/1 green Spike creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 green Spike"), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
