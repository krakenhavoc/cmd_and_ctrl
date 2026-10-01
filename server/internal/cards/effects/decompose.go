package effects

// Decompose — Sorcery {1}{B}:
//
//	"Exile up to three target cards from a single graveyard."
//
// The plainest card of the "from a single graveyard" family (#1807,
// ADR 0106 §5). The clause carries the sameness rule, so every pick
// comes from one player's graveyard: the announce gate refuses a set
// that reaches two, and the picker greys the other graveyards once the
// first card is chosen. A card that leaves in response is skipped and
// the rest are still exiled (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12c2b6bb-05f8-4291-b9a5-c4ddb7a69b9e",
		Name:         "Decompose",
		Completeness: CompletenessFull,
		Targets:      upToNCardsFromASingleGraveyard(3),
		OnResolve:    exileTargetCardsOnResolve,
	})
}
