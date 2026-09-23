package effects

// Wear Down — Sorcery {1}{G} (EDHREC rank 2403):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Destroy target artifact or enchantment. If the gift was
//	 promised, instead destroy two target artifacts and/or
//	 enchantments."
//
// A Naturalize that can become a two-for-one.
//
// Gift (CR 702.174, ADR 0089): the promise swaps the clause for its
// two-target twin (`.Instead(… .WithCount(2, 2))`, CR 702.174m), so
// the resolution simply destroys every target that is still legal —
// one for the unpromised cast, two for the promised one, and one if
// the other left in response (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "27905301-333e-4cdd-90cf-188159fcf8e9",
		Name:         "Wear Down",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		Gift: GiftACard().Instead(
			TargetPermanent("two target artifacts and/or enchantments", Or(Artifact(), Enchantment())).WithCount(2, 2)),
		OnResolve: destroyEachLegalTarget,
	})
}
