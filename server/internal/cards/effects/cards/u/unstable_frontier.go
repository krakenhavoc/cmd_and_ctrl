package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unstable Frontier — Land:
//
//	"{T}: Add {C}.
//	 {T}: Target land you control becomes the basic land type of your
//	 choice until end of turn."
//
// ADR 0109 §1 (#1881): CR 305.7 from a resolved ability. The type is
// chosen as the ability resolves (CR 608.2); until end of turn the land's
// land types are replaced by it (its other subtypes stay, CR 205.1a), it
// loses the abilities its rules text gives it and taps for the chosen
// type's colour (CR 305.6). It may target itself: it then loses both of
// its own abilities and becomes a basic-typed land for the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "495214b5-2eab-4fe4-8879-a30a57a67163",
		Name:         "Unstable Frontier",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target land you control becomes the basic land type of your choice until end of turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
			Effect:  effects.TargetLandBecomesUntilEOT("Unstable Frontier"),
		}},
	})
}
