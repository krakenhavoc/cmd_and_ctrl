package effects

// Urgent Necropsy — Instant {2}{B}{G} (Murders at Karlov Manor):
//
//	"As an additional cost to cast this spell, collect evidence X, where
//	 X is the total mana value of the permanents this spell targets.
//	 Destroy up to one target artifact, up to one target creature, up to
//	 one target enchantment, and up to one target planeswalker."
//
// The cost is CollectEvidenceFromTargetsCost (#5240): the targets are
// announced first (CR 601.2c), X is fixed from them as the total cost is
// determined (CR 601.2f) and the graveyard cards on exile_ids must reach
// it (CR 701.59a), so a cast that can't is refused (CR 118.3). No targets
// is X = 0. Each permanent counts once, even one chosen for two of the
// "target" words (an artifact creature may be both, CR 115.3: the card
// prints no "different"). The rulings say X is locked in once the targets
// are chosen; it is read once, at announce.
//
// The four "up to one" clauses are Teferi, Who Slows the Sunset's shape
// (Clauses + WithCount(0, 1)), and the destructions are one simultaneous
// event over the targets still legal (destroyEachLegalTarget), so a
// permanent chosen twice is destroyed once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "378a6203-0d47-4165-bdc2-43dfdfb4d15d",
		Name:           "Urgent Necropsy",
		Completeness:   CompletenessFull,
		AdditionalCost: CollectEvidenceFromTargetsCost(),
		Targets: Clauses(
			TargetPermanent("up to one target artifact", Artifact()).WithCount(0, 1),
			TargetCreature("up to one target creature").WithCount(0, 1),
			TargetPermanent("up to one target enchantment", Enchantment()).WithCount(0, 1),
			TargetPermanent("up to one target planeswalker", Planeswalker()).WithCount(0, 1),
		),
		OnResolve: destroyEachLegalTarget,
	})
}
