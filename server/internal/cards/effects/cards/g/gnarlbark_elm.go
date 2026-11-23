package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnarlbark Elm — Creature — Treefolk Warlock {2}{B}, 3/4 (#3987):
//
//	"This creature enters with two -1/-1 counters on it.
//	 {2}{B}, Remove two counters from this creature: Target creature
//	 gets -2/-2 until end of turn. Activate only as a sorcery."
//
// The counters it removes are of any kinds (RemoveCountersOfAnyKindsFromThis):
// usually its own two -1/-1 counters, which makes it a 3/4 again, but a
// +1/+1 counter it picked up pays as well. The activator chooses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a848ea46-347d-49c7-85c8-023cf836d0b3",
		Name:         "Gnarlbark Elm",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterMinusOne, 2, "Gnarlbark Elm: enters with two -1/-1 counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}, Remove two counters from this creature: Target creature gets -2/-2 until end of turn. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
			Cost:         effects.Plus(effects.ManaCost("{2}{B}"), effects.RemoveCountersOfAnyKindsFromThis(2)),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.VsPumpTheTarget(effects.NewContext(g, item), -2, -2, "Gnarlbark Elm — -2/-2")
			},
		}},
	})
}
