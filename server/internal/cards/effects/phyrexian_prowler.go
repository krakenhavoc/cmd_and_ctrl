package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Prowler — Creature — Phyrexian Zombie Mercenary {3}{B}, 3/3:
//
//	"Fading 3 (This creature enters with three fade counters on it. At
//	 the beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)
//	 Remove a fade counter from this creature: This creature gets +1/+1
//	 until end of turn."
//
// Fading is Fading/FadingUpkeep (fading.go, CR 702.32a). The pump's cost
// is the fade counter (RemoveAFadeCounterFromThis, CR 602.2b), so it can
// be activated once per counter, and the counters it spends shorten the
// Prowler's stay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "128ec451-97f6-4f01-8929-689fd4c45f40",
		Name:         "Phyrexian Prowler",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Fading(3)},
		Triggered:    []game.TriggeredAbility{FadingUpkeep()},
		Activated: []ActivatedAbility{{
			Label:   "Remove a fade counter from this creature: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    RemoveAFadeCounterFromThis(),
			Effect:  thisGetsUntilEndOfTurn(1, 1, "Phyrexian Prowler — +1/+1 until end of turn"),
		}},
	})
}
