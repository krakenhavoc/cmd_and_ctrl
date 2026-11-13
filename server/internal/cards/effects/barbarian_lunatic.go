package effects

// Barbarian Lunatic — Creature — Human Barbarian {2}{R}, 2/1 (Odyssey,
// #5613):
//
//	"{2}{R}, Sacrifice this creature: It deals 2 damage to target
//	 creature."
//
// The sacrifice is part of the cost (CR 602.2b), so "it" is the
// creature as it last existed on the battlefield (CR 113.7a): the
// damage comes from the Lunatic's last-known information.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12b5fbe5-a426-4b7e-bef0-abea71170e49",
		Name:         "Barbarian Lunatic",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice this creature: It deals 2 damage to target creature.",
			Cost:    Plus(ManaCost("{2}{R}"), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Effect:  DealDamageToTheTarget(2),
		}},
	})
}
