package effects

// Simic Initiate — Creature — Human Mutant {G}, 0/0 (Dissension, #5598):
//
//	"Graft 1 (This creature enters with a +1/+1 counter on it. Whenever
//	 another creature enters, you may move a +1/+1 counter from this
//	 creature onto it.)"
//
// Graft is disGraft: Cytoplast Manipulator's entry replacement plus
// cmdr2B23Graft (CR 702.58a).
//
// No simplification.
func init() {
	replacements, triggered := disGraft("Simic Initiate", 1)
	Register(Spec{
		OracleID:     "c7c37369-47f8-4027-a016-a2055307dd24",
		Name:         "Simic Initiate",
		Completeness: CompletenessFull,
		Replacements: replacements,
		Triggered:    triggered,
	})
}
