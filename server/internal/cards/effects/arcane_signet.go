package effects

// Arcane Signet — "{T}: Add one mana of any color in your
// commander's color identity." Color-identity-restricted mana rock.
//
// Fully implemented, same as Sol Ring: the S15 mana pipeline
// resolves the ability off-stack, and ActivateManaAbility narrows
// the five-colour pipe to the controller's commander's colour
// identity — which is exactly what this card's printed text asks
// for, and why it does NOT set IgnoreCommanderIdentity.
//
// The S14 note claiming the pipeline was still pending outlived its
// fix; corrected in the #338 stale-simplification sweep.
func init() {
	// Arcane Signet's produced string uses the full pipe set; the
	// engine narrows the options to the controller's commander's
	// color identity when the ability fires. The picker modal ends
	// up showing only the legal colors (W/U/G for a Bant deck, all
	// five for a Sisay deck, etc.).
	Register(Spec{
		OracleID:     "0bc7f093-bef0-4f1a-852c-4b75ebf54838",
		Name:         "Arcane Signet",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color in your commander's color identity",
		}},
	})
}
