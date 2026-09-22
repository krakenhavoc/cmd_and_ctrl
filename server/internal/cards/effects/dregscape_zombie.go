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
func init() {
	Register(Spec{
		OracleID:     "be9d1346-4416-4ade-ae84-7a4121e0bd12",
		Name:         "Dregscape Zombie",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"An unearthed creature bounced to hand goes to hand rather than being exiled; every other way it would leave the battlefield exiles it as printed.",
		},
		Activated: []ActivatedAbility{Unearth("{B}")},
	})
}
