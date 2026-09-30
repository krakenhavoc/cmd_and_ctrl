package effects

// Temporal Manipulation — Sorcery {3}{U}{U}:
//
//	"Take an extra turn after this one."
//
// CR 500.7 through TakeExtraTurn (ADR 0059 Decision 5, #753). Capture
// of Jingzhou is the same card under another name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b6ac99a-7548-4cad-9fe5-ea611618ab9e",
		Name:         "Temporal Manipulation",
		Completeness: CompletenessFull,
		OnResolve:    youTakeAnExtraTurn,
	})
}
