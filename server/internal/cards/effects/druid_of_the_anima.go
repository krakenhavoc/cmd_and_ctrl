package effects

// Druid of the Anima — Creature — Elf Druid {1}{G}, 1/1 (Foundations Jumpstart):
//
//	"{T}: Add {R}, {G}, or {W}."
//
// One three-colour pipe, so the activation is one picker (the tri-lands'
// shape); a summoning-sick Druid cannot tap for it (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d7933c4-4c74-492f-a9cd-50b339a4b29e",
		Name:         "Druid of the Anima",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|G|W}",
			Label:    "Add {R}, {G}, or {W}",
		}},
	})
}
