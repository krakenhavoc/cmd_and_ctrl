package effects

// Immaculate Magistrate — Creature — Elf Shaman {3}{G}, 2/2 (EDHREC
// rank 3420):
//
//	"{T}: Put a +1/+1 counter on target creature for each Elf you
//	 control."
//
// The Elf deck's Voltron button. A CR 602 tap ability — summoning
// sickness applies, as the engine enforces for every creature's tap
// cost — whose count is taken as the ability resolves: every Elf the
// controller controls, the Magistrate herself included (she is one),
// with a changeling counting through the effective subtypes. The
// counters go on through AddCounter, so a counter doubler applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6e38bd4-e6dc-400b-8e08-956726842dc4",
		Name:         "Immaculate Magistrate",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on target creature for each Elf you control.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  b32PutCountersPerElfOnChosenCreature,
		}},
	})
}
