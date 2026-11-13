package effects

// Unlikely Alliance — Enchantment {1}{W} (Alliances, #5634):
//
//	"{1}{W}: Target nonattacking, nonblocking creature gets +0/+2 until
//	 end of turn."
//
// A layer 7c boost (CR 613.4c) that ends at cleanup (CR 514.2). The
// target must be neither attacking nor blocking when the ability is
// activated and again as it resolves (CR 608.2b), so a creature that has
// since been declared as an attacker or blocker gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7ac18bc-ea63-4808-b709-92c1e69cf541",
		Name:         "Unlikely Alliance",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Target nonattacking, nonblocking creature gets +0/+2 until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 2)),
			Cost:    ManaCost("{1}{W}"),
			Targets: TargetCreature("target nonattacking, nonblocking creature", Not(AttackingOrBlocking())),
			Effect:  boostTargetUntilEOT("Unlikely Alliance — +0/+2 until end of turn", 0, 2),
		}},
	})
}
