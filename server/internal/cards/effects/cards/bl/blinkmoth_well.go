package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blinkmoth Well — Land (Mirrodin, #5606):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Tap target noncreature artifact."
//
// Witch's Clinic's colourless mana ability beside Auriok Transfixer's
// tapper, on a noncreature-artifact clause: an artifact creature (an
// animated Blinkmoth Nexus, an artifact land that is a creature) is not
// a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd3fd7cd-d8a1-48e2-861d-c7c40086dc75",
		Name:         "Blinkmoth Well",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Tap target noncreature artifact.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target noncreature artifact", effects.Artifact(), effects.Noncreature()),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
