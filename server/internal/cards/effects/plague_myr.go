package effects

// Plague Myr — Artifact Creature — Phyrexian Myr {2}, 1/1:
//
//	"Infect
//	 {T}: Add {C}."
//
// Infect is the engine's (ADR 0056, #748): the damage tail turns this
// creature's damage into -1/-1 counters on a creature and poison
// counters on a player. The keyword is declared here as well as
// stamped by the importer so fixtures and tokens carry it too. The mana
// ability is the whole of the rest of the card.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "2f328e05-5edf-4b21-9c2a-50dcf1e7b3ec",
		Name:            "Plague Myr",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
