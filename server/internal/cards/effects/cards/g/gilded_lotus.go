package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gilded Lotus — Artifact {5}:
//
//	"{T}: Add three mana of any one color."
//
// The canonical "N mana of any one color" (#742): ONE colour pick that
// adds three tokens of that colour, written in the produced-mana
// grammar as "{W3|U3|B3|R3|G3}". Before #742 each "{W|U|B|R|G}" slot
// was its own pick, so three slots would have let the player take
// three different colours — stronger than printed.
//
// The auto-tapper plans it (#779): the planner offers the solver one
// candidate per colour — three {U} slots, three {W} slots, and so on —
// and they are mutually exclusive, so the Lotus can fund {3}{U}{U}
// beside two Islands but never {W}{U} on its own. The colour the plan
// booked reaches the executor rather than being re-derived there.
//
// "Any one color", flatly — nothing about the commander's identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a02a9a7-39d9-4763-85d3-747a0540b60b",
		Name:         "Gilded Lotus",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.OneColorOfAmount(3),
			Label:    "Add three mana of any one color",
		}},
	})
}
