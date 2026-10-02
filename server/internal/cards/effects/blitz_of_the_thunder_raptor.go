package effects

// Blitz of the Thunder-Raptor — Instant {1}{R}:
//
//	"Blitz of the Thunder-Raptor deals damage to target creature or
//	 planeswalker equal to the number of instant and sorcery cards in
//	 your graveyard. If that creature or planeswalker would die this
//	 turn, exile it instead."
//
// The instant and sorcery cards are counted as the spell resolves
// (CR 608.2h), with the Blitz itself still on the stack. The replacement
// is the spell's (ADR 0108 §1): the target is marked even when the count
// is 0 or the damage is prevented.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d2db3f9c-26fe-487b-bbb7-8bc6f33456f1",
		Name:         "Blitz of the Thunder-Raptor",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve:    damageFirstTargetExileIfItDies(g2CardsInYourGraveyard(MatchInstantOrSorcery)),
	})
}
