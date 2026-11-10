package effects

// Teach by Example — Instant {U/R}{U/R} (Foundations):
//
//	"When you next cast an instant or sorcery spell this turn, copy that
//	 spell. You may choose new targets for the copy."
//
// Doublecast on an instant: the delayed trigger waits for the next
// instant or sorcery spell its controller casts this turn, whether or not
// that spell has targets, and puts a copy on the stack above it (a copy is
// never cast, CR 707.10). A modal spell's copy has the same modes, an X
// stays, and no additional cost is paid again: the shared CopySpell.
func init() {
	Register(Spec{
		OracleID:     "d85ba214-fa05-44ca-a7f9-ab2c5dc07962",
		Name:         "Teach by Example",
		Completeness: CompletenessFull,
		OnResolve:    copyYourNextInstantOrSorcery("Teach by Example — copy that spell"),
	})
}
