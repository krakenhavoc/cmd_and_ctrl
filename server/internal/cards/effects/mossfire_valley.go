package effects

// Mossfire Valley — Land (EDHREC top 1,000, #3654):
//
//	"{1}, {T}: Add {R}{G}."
//
// A filter land: no colourless option, one ability, {1} to filter into
// a fixed two colours. The Signet cycle's mana component
// (ManaAbilityCost.Mana) on a land, as Darkwater Catacombs is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23bbd091-f6ff-4514-97aa-42c08164b4eb",
		Name:         "Mossfire Valley",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{R}{G}",
			Label:    "{1}, {T}: Add {R}{G}",
		}},
	})
}
