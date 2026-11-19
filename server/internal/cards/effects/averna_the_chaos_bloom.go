package effects

// Averna, the Chaos Bloom — Legendary Creature — Elemental Shaman {G}{U}{R},
// 4/2 (Foundations Jumpstart):
//
//	"As you cascade, you may put a land card from among the exiled cards
//	 onto the battlefield tapped. (Do this after the last card is exiled but
//	 before deciding whether to cast a nonland card.)"
//
// A static of the permanent, read live by CascadeForEffect (Spec.
// AsYouCascadePutALand, like Teval's delve grant): once the exiling is done,
// and before the hit is offered, its controller may put one land from among
// the exiled cards onto the battlefield tapped. The unchosen lands go to the
// bottom with the rest. Applies to every cascade its controller performs,
// whoever's spell it is, and even when no nonland card was found.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:             "23695bc2-86d3-4d78-b374-9e83c4cc9cc4",
		Name:                 "Averna, the Chaos Bloom",
		Completeness:         CompletenessFull,
		AsYouCascadePutALand: true,
	})
}
