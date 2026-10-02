package effects

// Malicious Eclipse — Sorcery {1}{B}{B}:
//
//	"All creatures get -2/-2 until end of turn. If a creature an
//	 opponent controls would die this turn, exile it instead."
//
// The replacement's set is read as each creature would die (CR 611.2c,
// ADR 0108 §1): a creature is exiled if an opponent of the caster
// controls it then, whoever controlled it when the spell resolved.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "5beb8d6e-d3c1-46a5-8516-d6bf66413cff",
		Name:         "Malicious Eclipse",
		Completeness: CompletenessFull,
		OnResolve:    allCreaturesShrinkThenExileIfTheyDie(-2, true),
	})
}
