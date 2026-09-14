package effects

// Ulvenwald Tracker — Creature — Human Shaman {G}, 1/1 (EDHREC rank
// 2970):
//
//	"{1}{G}, {T}: Target creature you control fights another target
//	 creature."
//
// Green's repeatable fight. One CR 602 ability with a mana-and-tap
// cost (summoning sickness applies — the engine enforces it for a
// creature source) and two target slots read positionally: slot 0
// is the controller's creature, slot 1 the other creature, any
// controller. The fight is b10Fight — both powers read before either
// blow lands, so the two are simultaneous as CR 701.12 asks — and a
// creature that left in response fights nothing.
//
// DECLARED SIMPLIFICATION, weaker than printed — the Bite Down
// posture: the two slots have different clauses ("a creature you
// control", "another creature"), and a target spec is one predicate
// over every slot, so the spec is "a creature" over both and the
// "you control" half of slot 0 is checked at RESOLUTION rather than
// refused at announce (b28FightChosenTargets). A pair whose first
// creature is not yours makes the ability do nothing. Never
// stronger: no line exists that the printed card forbids.
func init() {
	Register(Spec{
		OracleID:     "63a37c08-7e38-4fe2-9a71-cc4604c4f831",
		Name:         "Ulvenwald Tracker",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The first target must be a creature you control, but that isn't checked until the ability resolves — a pair that doesn't fit makes it do nothing."},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}, {T}: Target creature you control fights another target creature.",
			Cost:    Plus(ManaCost("{1}{G}"), TapCost()),
			Targets: TargetCreature("target creature you control, then another target creature").WithCount(2, 2),
			Effect:  b28FightChosenTargets,
		}},
	})
}
