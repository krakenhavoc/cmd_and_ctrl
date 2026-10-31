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
// blow lands, so the two are simultaneous as CR 701.14 asks — and a
// creature that left in response fights nothing.
//
// Two target CLAUSES (Clauses, the Bite Down shape): slot 0 carries
// "you control" and slot 1 "another" (Distinct), each enforced at
// announce (CR 601.2c) and re-checked per slot at resolution
// (CR 608.2b), so a pair that does not fit is refused when activated
// rather than doing nothing.
func init() {
	Register(Spec{
		OracleID:     "63a37c08-7e38-4fe2-9a71-cc4604c4f831",
		Name:         "Ulvenwald Tracker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}{G}, {T}: Target creature you control fights another target creature.",
			Cost:  Plus(ManaCost("{1}{G}"), TapCost()),
			Targets: Clauses(
				TargetCreature("target creature you control", YouControl()),
				Distinct(TargetCreature("another target creature")),
			),
			Effect: slotZeroFightsSlotOne,
		}},
	})
}
