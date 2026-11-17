package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tamiyo's Immobilizer — Artifact {3}{U} (Phyrexia: All Will Be One):
//
//	"This artifact enters with four oil counters on it.
//	 {T}, Remove an oil counter from this artifact: Tap target artifact or
//	 creature."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2c7a9c2-28f8-4cf0-a042-420f6a4ba8bd",
		Name:         "Tamiyo's Immobilizer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b10EntersWithCounters(oilCounter, 4, "Tamiyo's Immobilizer: enters with four oil counters on it")},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove an oil counter from this artifact: Tap target artifact or creature.",
			Cost:    Plus(TapCost(), RemoveCountersFromThis(oilCounter, 1)),
			Targets: TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
			Effect:  tapChosenPermanent,
		}},
	})
}
