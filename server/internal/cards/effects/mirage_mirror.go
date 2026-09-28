package effects

// Mirage Mirror — Artifact {3}:
//
//	"{2}: This artifact becomes a copy of target artifact, creature,
//	 enchantment, or land until end of turn."
//
// The card #1593 names first: a duration copy on a permanent already
// on the battlefield (become_copy.go). While it is the copy it has the
// copied card's text and none of its own — including this ability, so
// a Mirage Mirror that has become a Grizzly Bears cannot activate it
// again until cleanup puts it back. A second activation in response,
// while it is still a Mirror, is legal, and the later copy wins
// (CR 613.7).
//
// Copying a creature does not make it summoning-sick on its own: the
// artifact has been under its controller's control since it entered,
// so it may attack if that was before this turn (CR 302.6).
func init() {
	Register(Spec{
		OracleID:     "6a84eda1-7e72-4b5a-849f-b6e177565aeb",
		Name:         "Mirage Mirror",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{2}: This artifact becomes a copy of target artifact, creature, enchantment, or land until end of turn.",
			Cost:  ManaCost("{2}"),
			Targets: TargetPermanent("target artifact, creature, enchantment, or land",
				Or(Artifact(), Creature(), Enchantment(), Land())),
			Effect: selfBecomesCopyOfTarget("Mirage Mirror — becomes a copy until end of turn"),
		}},
	})
}
