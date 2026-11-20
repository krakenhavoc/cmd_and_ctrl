package effects

// Tempered Veteran — Creature — Human Knight {1}{W}, 1/2 (Core Set 2021, #5511):
//
//	"{W}, {T}: Put a +1/+1 counter on target creature with a +1/+1 counter
//	 on it.
//	 {4}{W}{W}, {T}: Put a +1/+1 counter on target creature."
//
// Dragon Blood's counter (PlusOneCounterOnTheTarget) twice: the cheap row
// takes Helium Squirter's "with a +1/+1 counter on it" clause, judged at
// announce and again as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c19f3f00-2c7b-4e85-a034-e65760d9cc13",
		Name:         "Tempered Veteran",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{W}, {T}: Put a +1/+1 counter on target creature with a +1/+1 counter on it.",
				Cost:    Plus(ManaCost("{W}"), TapCost()),
				Targets: TargetCreature("target creature with a +1/+1 counter on it", Not(b13WithoutPlusOneCounter())),
				Effect:  PlusOneCounterOnTheTarget,
			},
			{
				Label:   "{4}{W}{W}, {T}: Put a +1/+1 counter on target creature.",
				Cost:    Plus(ManaCost("{4}{W}{W}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  PlusOneCounterOnTheTarget,
			},
		},
	})
}
