package effects

// Mystic Gate — Land (EDHREC rank 871):
//
//	"{T}: Add {C}.
//	 {W/U}, {T}: Add {W}{W}, {W}{U}, or {U}{U}."
//
// The first Shadowmoor filter land in the catalog, and the shape the
// roadmap's batch 02 skipped five of — writable since #356 gave
// ManaAbilityCost a mana component. The hybrid {W/U} parses as a
// requirement either colour satisfies, paid from the pool (the
// engine has no auto-tap into a mana ability's cost).
//
// The three-way OUTPUT — the question batch 02 left open — is two
// independent pipe slots: "{W|U}{W|U}" resolves to WW, WU, UW or UU,
// exactly the printed three outcomes, one colour pick per slot.
// Printed colours, so no commander-identity narrowing; the painless
// {C} ability sits first so the auto-tapper reaches for it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e9f5feb2-2c1a-46ce-885a-4f378d7d10af",
		Name:         "Mystic Gate",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{W/U}"},
				Produced: "{W|U}{W|U}",
				Label:    "{W/U}, {T}: Add {W}{W}, {W}{U}, or {U}{U}",
			},
		},
	})
}
