package effects

// Cytospawn Shambler — Creature — Elemental Mutant {6}{G}, 0/0 (Dissension, #5598):
//
//	"Graft 6 (This creature enters with six +1/+1 counters on it. Whenever
//	 another creature enters, you may move a +1/+1 counter from this
//	 creature onto it.)
//	 {G}: Target creature with a +1/+1 counter on it gains trample until
//	 end of turn."
//
// Graft is disGraft: Cytoplast Manipulator's entry replacement plus
// cmdr2B23Graft (CR 702.58a).
// The grant is disGraftGrantRow with trample.
//
// No simplification.
func init() {
	replacements, triggered := disGraft("Cytospawn Shambler", 6)
	Register(Spec{
		OracleID:     "514c6831-63ac-4fc0-8950-526813591710",
		Name:         "Cytospawn Shambler",
		Completeness: CompletenessFull,
		Replacements: replacements,
		Triggered:    triggered,
		Activated:    []ActivatedAbility{disGraftGrantRow("{G}", "trample", "Cytospawn Shambler")},
	})
}
