package effects

// Promising Vein — Land — Cave (EDHREC rank 3191):
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Search your library for a basic
//	 land card, put it onto the battlefield tapped, then shuffle."
//
// An Evolving Wilds that taps for colourless while it waits and
// costs {1} to crack. The search is the shared fetch shape with the
// basic read off the Basic supertype, so a snow-covered basic
// qualifies; the fetched land enters tapped through the search's
// own entry flag. No sorcery-speed gate, and a land is never
// summoning sick, so it can be cracked the turn it is played.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "861eb7d7-7616-4620-a4fd-4b8c3bf00dd1",
		Name:         "Promising Vein",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:  "{1}, {T}, Sacrifice Promising Vein: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Cost:   Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Effect: b30FetchBasicTapped,
		}},
	})
}
