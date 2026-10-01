package effects

// Scarab Feast — Instant {B}:
//
//	"Exile up to three target cards from a single graveyard.
//	 Cycling {B} ({B}, Discard this card: Draw a card.)"
//
// Decompose at instant speed with cycling (#1807, ADR 0106 §5). The
// clause and the body are Decompose's; cycling is the shared
// constructor, an activated ability from the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b9043d3-a1c6-4f49-9c49-ef78bbfbd4ac",
		Name:         "Scarab Feast",
		Completeness: CompletenessFull,
		Targets:      upToNCardsFromASingleGraveyard(3),
		OnResolve:    exileTargetCardsOnResolve,
		Activated:    []ActivatedAbility{Cycling("{B}")},
	})
}
