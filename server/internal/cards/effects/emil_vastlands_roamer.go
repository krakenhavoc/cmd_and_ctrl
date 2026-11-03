package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emil, Vastlands Roamer — Legendary Creature — Elf Druid {2}{G}, 3/3
// (Secrets of Strixhaven, #4456):
//
//	"Creatures you control with +1/+1 counters on them have trample.
//	 {4}{G}, {T}: Create a 0/0 green and blue Fractal creature token. Put X
//	 +1/+1 counters on it, where X is the number of differently named
//	 lands you control."
//
// The trample is a layer-6 keyword grant over its controller's creatures
// that carry a +1/+1 counter (Great Ugly-Looking Goblin's shape), Emil
// included, read live off the counter map. The ability creates the Fractal
// first and then puts the counters on it, as printed (Deekah's shape,
// socB8FractalWithCounters); X is counted as the ability resolves, each land
// once however many share its name (the ruling), so five Forests are one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f800ebc-d6e4-4bfc-bebe-8baad51efa19",
		Name:         "Emil, Vastlands Roamer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{b16GrantKeywords(b42CreatureYouControlWithAPlusOneCounter, "trample")},
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}, {T}: Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it, where X is the number of differently named lands you control.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Cost:    Plus(ManaCost("{4}{G}"), TapCost()),
			Effect:  socB8FractalWithCounters("emil-vastlands-roamer/fractal", socB23DifferentlyNamedLands),
		}},
	})
}
