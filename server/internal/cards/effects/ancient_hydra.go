package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Hydra — Creature — Hydra {4}{R}, 5/1:
//
//	"Fading 5 (This creature enters with five fade counters on it. At the
//	 beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)
//	 {1}, Remove a fade counter from this creature: It deals 1 damage to
//	 any target."
//
// Fading is Fading/FadingUpkeep (fading.go, CR 702.32a). The ability is
// Deathbringer Thoctar's ping with {1} added to the cost; the counter is
// paid as it is activated (RemoveAFadeCounterFromThis, CR 602.2b), so
// the Hydra can spend its counters in response to removal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e1d3722-9bd1-4ab7-b4a5-9265fca3aba9",
		Name:         "Ancient Hydra",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Fading(5)},
		Triggered:    []game.TriggeredAbility{FadingUpkeep()},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Remove a fade counter from this creature: It deals 1 damage to any target.",
			Cost:    Plus(ManaCost("{1}"), RemoveAFadeCounterFromThis()),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  DealDamageToTheTarget(1),
		}},
	})
}
