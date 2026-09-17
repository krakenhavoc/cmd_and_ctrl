package effects

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
// The auto-tapper plans around a one-colour-N-mana source (its
// one-slot-one-mana model can't promise the tokens share a colour), so
// the Lotus is tapped by hand and the floated mana is spent by the
// cast like any other. That is a planner convenience, not a card
// simplification.
//
// "Any one color", flatly — nothing about the commander's identity.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a02a9a7-39d9-4763-85d3-747a0540b60b",
		Name:         "Gilded Lotus",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                OneColorOfAmount(3),
			Label:                   "Add three mana of any one color",
			IgnoreCommanderIdentity: true,
		}},
	})
}
