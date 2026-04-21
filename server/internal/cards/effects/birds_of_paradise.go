package effects

// Birds of Paradise — 0/1 Creature — Bird with "Flying" and
// "{T}: Add one mana of any color."
//
// S14 sandbox: flying keyword is cosmetic (keyword enforcement is
// S18); the mana ability is activated and resolves off-engine
// today (S19). Registered as a vanilla catalog entry so it carries
// the AUTO badge — the sprint plan's explicit example of "vanilla
// creature with no ETB" in spec.go.
func init() {
	Register(Spec{
		OracleID: "d3a0b660-358c-41bd-9cd2-41fbf3491b1a",
		Name:     "Birds of Paradise",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
