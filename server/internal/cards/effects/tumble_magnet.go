package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tumble Magnet — Artifact {3} (Scars of Mirrodin, #5577):
//
//	"This artifact enters with three charge counters on it.
//	 {T}, Remove a charge counter from this artifact: Tap target artifact or
//	 creature."
//
// Tamiyo's Immobilizer's shape with three charge counters: the counters
// are a self-replacement on its entry (CR 614.1c), and removing one is
// part of the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0481adcd-b66d-432e-80ba-3a4718bb3a7d",
		Name:         "Tumble Magnet",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterCharge, 3, "Tumble Magnet: enters with three charge counters"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove a charge counter from this artifact: Tap target artifact or creature.",
			Cost:    Plus(TapCost(), RemoveCountersFromThis(game.CounterCharge, 1)),
			Targets: TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
			Effect:  tapChosenPermanent,
		}},
	})
}
