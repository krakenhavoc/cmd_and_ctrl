package effects

// Maraleaf Pixie — Creature — Faerie {G}{U}, 2/2 (Throne of Eldraine, #5516):
//
//	"Flying
//	 {T}: Add {G} or {U}."
//
// A mana ability (CR 605.1a): it does not use the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "013e4281-4e3e-47c2-ac9a-f570668bcb16",
		Name:            "Maraleaf Pixie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
	})
}
