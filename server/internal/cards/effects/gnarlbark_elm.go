package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a848ea46-347d-49c7-85c8-023cf836d0b3",
		Name:         "Gnarlbark Elm",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 2, "Gnarlbark Elm: enters with two -1/-1 counters"),
		},
		Activated: []ActivatedAbility{{
			Label:        "{2}{B}, Remove two counters from this creature: Target creature gets -2/-2 until end of turn. Activate only as a sorcery.",
			Purpose:      ForTargets(HarmUntilEOT(0, -2, -2)),
			Cost:         Plus(ManaCost("{2}{B}"), RemoveCountersOfAnyKindsFromThis(2)),
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return vsPumpTheTarget(NewContext(g, item), -2, -2, "Gnarlbark Elm — -2/-2")
			},
		}},
	})
}
