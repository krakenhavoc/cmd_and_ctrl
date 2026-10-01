package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Experiment One — Creature — Human Ooze {G}, 1/1:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 Remove two +1/+1 counters from this creature: Regenerate it. (The
//	 next time this creature would be destroyed this turn, instead tap
//	 it, remove it from combat, and heal all damage on it.)"
//
// Evolve is the engine's keyword trigger (game/evolve.go, #1805). The
// regeneration is a counter-removal cost (ADR 0020 addendum) paying for
// an ordinary CR 701.19a regeneration shield (Regenerate).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8520ee67-439b-4f15-838b-420bacbb8b13",
		Name:            "Experiment One",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		Activated: []ActivatedAbility{{
			Label: "Remove two +1/+1 counters from this creature: Regenerate it.",
			Cost:  RemoveCountersFromThis(game.CounterPlusOne, 2),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
