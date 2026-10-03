package effects

// Elvish Scout — Creature — Elf Scout {G}, 1/1:
//
//	"{G}, {T}: Untap target attacking creature you control. Prevent all
//	 combat damage that would be dealt to and dealt by it this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): Ebony Horse's ability on a
// creature, so the {T} needs it to have been under your control since
// your turn began (CR 302.6).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d39a492e-a599-4283-bfc1-8e99065a9303",
		Name:         "Elvish Scout",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{untapAttackerToAndByRow(
			"{G}, {T}: Untap target attacking creature you control. Prevent all combat damage that would be dealt to and dealt by it this turn.",
			Plus(ManaCost("{G}"), TapCost()),
			TargetCreature("target attacking creature you control", AttackingCreature(), YouControl()))},
	})
}
