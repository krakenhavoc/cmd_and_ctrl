package effects

// Sultai Devotee — Creature — Zombie Snake Druid {1}{G}, 2/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"Deathtouch
//	 {1}: Add {B}, {G}, or {U}. Activate only once each turn."
//
// Abzan Devotee's filter with the Sultai colours: a Signet-style
// `ManaAbilityCost.Mana` with no tap, limited by the declarative
// OncePerTurn bit, counted per object (a Devotee that leaves and comes
// back is a new object with a fresh use, CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ee7b3fc0-6128-4bf6-b3e7-7371e6f93c4b",
		Name:            "Sultai Devotee",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		ManaAbilities: []ManaAbility{{
			Cost:        ManaAbilityCost{Mana: "{1}"},
			Produced:    "{B|G|U}",
			Label:       "{1}: Add {B}, {G}, or {U}",
			OncePerTurn: true,
		}},
	})
}
