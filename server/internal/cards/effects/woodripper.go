package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Woodripper — Creature — Beast {3}{G}{G}, 4/6:
//
//	"Fading 3 (This creature enters with three fade counters on it. At
//	 the beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)
//	 {1}, Remove a fade counter from this creature: Destroy target
//	 artifact."
//
// Fading is Fading/FadingUpkeep (fading.go, CR 702.32a). The ability is
// District Mascot's artifact kill with one fade counter for its counter
// cost (RemoveAFadeCounterFromThis, CR 602.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0653734-9eb7-48c6-9042-65ffdd50fd2a",
		Name:         "Woodripper",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Fading(3)},
		Triggered:    []game.TriggeredAbility{FadingUpkeep()},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Remove a fade counter from this creature: Destroy target artifact.",
			Cost:    Plus(ManaCost("{1}"), RemoveAFadeCounterFromThis()),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
