package effects

// Shred Memory — Instant {1}{B}:
//
//	"Exile up to four target cards from a single graveyard.
//	 Transmute {1}{B}{B} ({1}{B}{B}, Discard this card: Search your
//	 library for a card with the same mana value as this card, reveal
//	 it, put it into your hand, then shuffle. Transmute only as a
//	 sorcery.)"
//
// #1807, ADR 0106 §5. The spell is Decompose with a fourth target;
// transmute is the new constructor in transmute.go, an activated
// ability from the hand that finds another two-mana-value card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16cae855-d93a-445d-9cac-dbc5a7c18cb1",
		Name:         "Shred Memory",
		Completeness: CompletenessFull,
		Targets:      upToNCardsFromASingleGraveyard(4),
		OnResolve:    exileTargetCardsOnResolve,
		Activated:    []ActivatedAbility{Transmute("{1}{B}{B}")},
	})
}
