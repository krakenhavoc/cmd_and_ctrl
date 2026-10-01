package effects

// Rapid Decay — Instant {1}{B}:
//
//	"Exile up to three target cards from a single graveyard.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Scarab Feast with a different cycling cost (#1807, ADR 0106 §5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c769dfc-79be-4fcb-840c-b72cafdd0a23",
		Name:         "Rapid Decay",
		Completeness: CompletenessFull,
		Targets:      upToNCardsFromASingleGraveyard(3),
		OnResolve:    exileTargetCardsOnResolve,
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
