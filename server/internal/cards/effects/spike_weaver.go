package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spike Weaver — Creature — Spike {2}{G}{G}, 0/0 (Stronghold, Mystery Booster 2):
//
//	"This creature enters with three +1/+1 counters on it.
//	 {2}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter
//	 on target creature.
//	 {1}, Remove a +1/+1 counter from this creature: Prevent all combat
//	 damage that would be dealt this turn."
//
// The Spike cycle's shared entry replacement and counter move
// (spike_counters.go); the second ability is Fog's shield, paid with a
// counter and {1} (CR 118.3), so it works at instant speed and on a
// summoning-sick Weaver.
//
// No simplification.
func init() {
	const name = "Spike Weaver"
	Register(Spec{
		OracleID:     "7776b2dd-5795-458d-81da-54fed22d5ee2",
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{spikeEntersWith(name, "three", 3)},
		Activated: []ActivatedAbility{
			spikeMovesACounter(),
			{
				Label:   "{1}, Remove a +1/+1 counter from this creature: Prevent all combat damage that would be dealt this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    Plus(ManaCost("{1}"), RemoveCountersFromThis(game.CounterPlusOne, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return PreventAllCombatDamageThisTurn{Label: "Spike Weaver: prevent combat damage"}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
