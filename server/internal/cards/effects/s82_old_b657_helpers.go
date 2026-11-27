package effects

// S82OldB657Cameo is a Cameo from Invasion (#5616): "{T}: Add {a} or {b}."
// on a coloured-pair artifact, DualManaAbility's two-colour choice.
func S82OldB657Cameo(oracleID, name, a, b string) Spec {
	return Spec{
		OracleID:      oracleID,
		Name:          name,
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{DualManaAbility(a, b)},
	}
}

// S82OldB657Attendant is an Invasion Attendant (#5616):
// "{1}, Sacrifice this creature: Add <three colours>." A mana ability
// (CR 605.1a) whose cost pays {1} and sacrifices the Golem.
func S82OldB657Attendant(oracleID, name, produced string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Mana: "{1}", Sacrifice: true},
			Produced: produced,
			Label:    "{1}, Sacrifice this creature: Add " + produced,
		}},
	}
}
