package effects

// Dregscape Zombie — Creature — Zombie {1}{B}, 2/1:
//
//	"Unearth {B} ({B}: Return this card from your graveyard to the
//	 battlefield. It gains haste. Exile it at the beginning of the
//	 next end step or if it would leave the battlefield. Unearth only
//	 as a sorcery.)"
//
// The whole card is the keyword, which is why it is #1221's unearth
// proof: a 2/1 body with no other text, so a test that sees it come
// back is seeing the keyword and nothing else.
//
// Unearth is an ACTIVATED ability that functions from the graveyard
// (CR 702.82a), so this Spec is one `Unearth("{B}")` entry and the
// zone dimension does the rest — the same shape a Triome's cycling
// has one zone over. See unearth.go.
//
// Its one caveat — a bounce skipping the redirect — went stale with
// #539, which routed BounceToHandForEffect through the shared exit
// primitive and its CR 614 window; ADR 0041 tier 3b's rework of the
// redirect pinned it (TestUnearthedCreatureBouncedGoesToExile).
func init() {
	Register(Spec{
		OracleID:     "be9d1346-4416-4ade-ae84-7a4121e0bd12",
		Name:         "Dregscape Zombie",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Unearth("{B}")},
	})
}
