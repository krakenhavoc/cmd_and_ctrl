package effects

// Noxious Newt — Creature — Salamander {1}{G}, 1/2 (Secrets of Strixhaven,
// #4456):
//
//	"Deathtouch
//	 {T}: Add {G}."
//
// Deathtouch is a printed keyword; the second line is an ordinary tap mana
// ability (Ominous Asylum's shape), so it obeys summoning sickness (CR
// 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1232c461-643c-40a3-901e-09b665f817da",
		Name:            "Noxious Newt",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
