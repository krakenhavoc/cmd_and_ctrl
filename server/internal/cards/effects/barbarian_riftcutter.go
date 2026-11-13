package effects

// Barbarian Riftcutter — Creature — Human Barbarian {4}{R}, 3/3
// (Ravnica: City of Guilds, #5600):
//
//	"{R}, Sacrifice this creature: Destroy target land."
//
// Strip Mine's destroy on a creature's sacrifice: the creature is paid as
// the cost, so the ability resolves with its source already gone, and a
// land that left in response is not destroyed (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e9612e4c-1527-4255-9f2f-b8d8cee65cc6",
		Name:         "Barbarian Riftcutter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: Destroy target land.",
			Cost:    Plus(ManaCost("{R}"), SacrificeThis()),
			Targets: TargetPermanent("target land", Land()),
			Effect:  destroyChosenPermanent,
		}},
	})
}
