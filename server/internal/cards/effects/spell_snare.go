package effects

// Spell Snare — Instant {U} (Lorwyn Eclipsed, #4458):
//
//	"Counter target spell with mana value 2."
//
// Mental Misstep's clause at two: the mana value is read off the spell as
// it sits on the stack (b03ManaValueIs), so {X} counts at its chosen value
// (CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38877c68-b69b-4ee1-8e8f-5137a184a8f7",
		Name:         "Spell Snare",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell with mana value 2", b03ManaValueIs(2)),
		OnResolve:    counterTheTargetSpell,
	})
}
