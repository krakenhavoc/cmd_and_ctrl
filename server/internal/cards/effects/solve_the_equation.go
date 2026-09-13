package effects

// Solve the Equation — Sorcery, {2}{U} (EDHREC rank 932):
//
//	"Search your library for an instant or sorcery card, reveal it,
//	 put it into your hand, then shuffle."
//
// The blue Idyllic Tutor: the same reveal-and-tutor body with the
// instant-or-sorcery predicate. The search is the S22 chooser, so the
// caster picks; the revealed card is known to the table.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f02682f0-26c0-4032-9df3-273b6a45d0a8",
		Name:         "Solve the Equation",
		Completeness: CompletenessFull,
		OnResolve:    b06TutorToHand("Solve the Equation — an instant or sorcery card, revealed, to your hand", b08IsInstantOrSorceryCard),
	})
}
