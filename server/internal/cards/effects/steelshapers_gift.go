package effects

// Steelshaper's Gift — Sorcery {W} (EDHREC rank 1036):
//
//	"Search your library for an Equipment card, reveal that card, put
//	 it into your hand, then shuffle."
//
// The one-mana Equipment tutor. Idyllic Tutor's body (b06TutorToHand)
// with the Equipment subtype as the filter; the searcher picks, the
// pick is revealed, the library shuffles.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9abda7e-6ca2-42ea-ab24-c542e57014f1",
		Name:         "Steelshaper's Gift",
		Completeness: CompletenessFull,
		OnResolve:    b06TutorToHand("Steelshaper's Gift — an Equipment card, revealed, to hand", b09IsEquipmentCard),
	})
}
