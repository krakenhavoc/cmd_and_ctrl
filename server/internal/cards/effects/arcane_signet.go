package effects

// Arcane Signet — "{T}: Add one mana of any color in your
// commander's color identity." Color-identity-restricted mana rock.
//
// S14 sandbox: same posture as Sol Ring — activated mana abilities
// land with the S19 pipeline. Registered here as a vanilla catalog
// entry so the AUTO badge reads consistently across Commander
// staples, and the future activated-ability wiring has a Spec to
// attach to.
func init() {
	// Arcane Signet's produced string uses the full pipe set; the
	// engine narrows the options to the controller's commander's
	// color identity when the ability fires. The picker modal ends
	// up showing only the legal colors (W/U/G for a Bant deck, all
	// five for a Sisay deck, etc.).
	Register(Spec{
		OracleID: "0bc7f093-bef0-4f1a-852c-4b75ebf54838",
		Name:     "Arcane Signet",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color in your commander's color identity",
		}},
	})
}
