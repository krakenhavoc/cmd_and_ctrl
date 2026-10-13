package effects

// Book Burning — Sorcery {1}{R}:
//
//	"Any player may have Book Burning deal 6 damage to them. If no one
//	 does, target player mills six cards."
//
// ADR 0150 §3 (#1903): Browbeat's offer with a mill
// (anyPlayerMayTakeDamageFromThis). A target gone by the time nobody
// pays mills nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "644a8946-4890-4c2b-978f-02eedc95b5ad",
		Name:         "Book Burning",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: anyPlayerMayTakeDamageFromThis("Book Burning", 6, "the mill", func(ctx *Context) error {
			who, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return MillCards{Player: who, N: 6}.Apply(ctx)
		}),
	})
}
