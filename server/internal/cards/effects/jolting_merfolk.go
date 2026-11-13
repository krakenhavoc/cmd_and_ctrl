package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jolting Merfolk — Creature — Merfolk {2}{U}{U}, 2/2:
//
//	"Fading 4 (This creature enters with four fade counters on it. At the
//	 beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)
//	 Remove a fade counter from this creature: Tap target creature."
//
// Fading is Fading/FadingUpkeep (fading.go, CR 702.32a). The tapper's
// cost is the fade counter (RemoveAFadeCounterFromThis, CR 602.2b) and
// its body is the shared "Tap target <permanent>." (tapChosenPermanent).
// Spending every counter leaves the upkeep trigger nothing to remove, so
// it sacrifices the Merfolk at the next upkeep.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b71aeee-a4a8-46d3-9973-06db14d8bf6c",
		Name:         "Jolting Merfolk",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Fading(4)},
		Triggered:    []game.TriggeredAbility{FadingUpkeep()},
		Activated: []ActivatedAbility{{
			Label:   "Remove a fade counter from this creature: Tap target creature.",
			Cost:    RemoveAFadeCounterFromThis(),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
