package effects

// Maze of Ith — Land:
//
//	"{T}: Untap target attacking creature. Prevent all combat damage that
//	 would be dealt to and dealt by that creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the untap, then ONE prevention
// effect over both directions — a preventFromSource record pinned to
// the creature with Mod.AndDealtBy, so its combat damage and the combat
// damage dealt to it are prevented by the same effect (CR 614.5's one
// opportunity per event, CR 615.13's one application per instance).
// Non-combat damage to and from it is dealt. A creature that leaves and
// returns is a new object the shield no longer covers (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "38a12bd7-4394-44a8-91a0-6a4ff7fa4f71",
		Name:         "Maze of Ith",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{untapAttackerToAndByRow(
			"{T}: Untap target attacking creature. Prevent all combat damage that would be dealt to and dealt by that creature this turn.",
			TapCost(), TargetCreature("target attacking creature", AttackingCreature()))},
	})
}
