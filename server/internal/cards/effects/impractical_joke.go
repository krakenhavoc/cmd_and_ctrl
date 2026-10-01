package effects

// Impractical Joke — Sorcery {R}:
//
//	"Damage can't be prevented this turn. Impractical Joke deals 3
//	 damage to up to one target creature or planeswalker."
//
// The turn grant (ADR 0107 §5) begins first. "Up to one": with no
// target chosen the spell still resolves and the grant still happens.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "c8da73ed-2834-4094-8723-a4ad75660290",
		Name:         "Impractical Joke",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("up to one target creature or planeswalker", Or(Creature(), Planeswalker())).WithCount(0, 1),
		OnResolve:    damageCantBePreventedThen(damageToFirstTarget(3)),
	})
}
