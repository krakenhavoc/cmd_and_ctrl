package effects

// Fellwar Stone — "{T}: Add one mana of any color that a land an
// opponent controls could produce."
//
// Sandbox simplification: the produced set is the full five-colour
// pipe rather than the intersection of what opponents' lands could
// actually produce. Narrowing needs a per-activation scan of every
// opponent's battlefield lands and their mana abilities; the engine
// has an equivalent hook only for commander colour identity
// (Arcane Signet). Documented rather than silently over-permissive:
// in a real game this can offer a colour no opponent could make.
// Follow-up work is a ManaAbility-level filter analogous to
// commanderIdentityFor.
func init() {
	Register(Spec{
		OracleID: "95560508-7ac9-4be9-8a3f-3c7d5b52807b",
		Name:     "Fellwar Stone",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color an opponent's land could produce",
		}},
	})
}
