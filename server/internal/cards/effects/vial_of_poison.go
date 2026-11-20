package effects

// Vial of Poison — Artifact {1} (Magic 2014, #5563):
//
//	"{1}, Sacrifice this artifact: Target creature gains deathtouch until end of turn. (Any amount of damage it deals to a creature is enough to destroy it.)"
//
// Trostani's deathtouch grant from a one-shot artifact: the Vial is
// sacrificed as the cost, and the grant still resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d51077f-4ce5-475a-9e65-e9076d779825",
		Name:         "Vial of Poison",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this artifact: Target creature gains deathtouch until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch")),
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Effect:  jmpB7FirstLegalTargetGains("Vial of Poison — deathtouch until end of turn", "deathtouch"),
		}},
	})
}
