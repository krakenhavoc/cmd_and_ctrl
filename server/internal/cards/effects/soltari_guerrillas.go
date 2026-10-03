package effects

// Soltari Guerrillas — Creature — Soltari Soldier {2}{R}{W}, 3/2:
//
//	"Shadow
//	 {0}: The next time this creature would deal combat damage to an
//	 opponent this turn, it deals that damage to target creature instead."
//
// ADR 0108 §9 (#1905): a "next time" redirection of this creature's
// combat damage to any opponent — read as the damage would be dealt, so
// it is whoever it is attacking then (the ruling) — to the target.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "37da03c2-c03e-453c-8fbf-e1039faceb8c",
		Name:         "Soltari Guerrillas",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{0}: The next time this creature would deal combat damage to an opponent this turn, it deals that damage to target creature instead.",
				ManaCost("{0}"), TargetCreature("target creature"),
				RedirectDamage{FromThis: true, Opponents: true, CombatOnly: true, Next: true, To: RedirectToClause(0)}),
		},
	})
}
