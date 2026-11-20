package effects

// Pestilent Souleater — Artifact Creature — Phyrexian Insect {5}, 3/3 (New Phyrexia, #5575):
//
//	"{B/P}: This creature gains infect until end of turn. ({B/P} can be paid with either {B} or 2 life. A creature with infect deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)"
//
// Vector Asp's row (somB1SelfGainsUntilEOT) for {B/P}; the Phyrexian
// symbol is paid with {B} or 2 life (CR 107.4f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4779671-2a01-4eef-847c-ede36091fc4a",
		Name:         "Pestilent Souleater",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{somB1SelfGainsUntilEOT("Pestilent Souleater", "{B/P}", "infect")},
	})
}
