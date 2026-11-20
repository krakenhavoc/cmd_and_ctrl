package effects

// Plaxcaster Frogling — Creature — Frog Mutant {1}{G}{U}, 0/0 (Dissension, #5598):
//
//	"Graft 3 (This creature enters with three +1/+1 counters on it. Whenever
//	 another creature enters, you may move a +1/+1 counter from this
//	 creature onto it.)
//	 {2}: Target creature with a +1/+1 counter on it gains shroud until
//	 end of turn. (It can't be the target of spells or abilities.)"
//
// Graft is disGraft: Cytoplast Manipulator's entry replacement plus
// cmdr2B23Graft (CR 702.58a).
// The grant is disGraftGrantRow with shroud.
//
// No simplification.
func init() {
	replacements, triggered := disGraft("Plaxcaster Frogling", 3)
	Register(Spec{
		OracleID:     "8c1ea1a5-5966-4d14-8361-ffc6a8d1bd79",
		Name:         "Plaxcaster Frogling",
		Completeness: CompletenessFull,
		Replacements: replacements,
		Triggered:    triggered,
		Activated:    []ActivatedAbility{disGraftGrantRow("{2}", "shroud", "Plaxcaster Frogling")},
	})
}
