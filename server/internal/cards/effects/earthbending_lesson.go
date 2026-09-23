package effects

// Earthbending Lesson — Sorcery — Lesson {3}{G}:
//
//	"Earthbend 4. (Target land you control becomes a 0/0 creature with
//	 haste that's still a land. Put four +1/+1 counters on it. When it
//	 dies or is exiled, return it to the battlefield tapped.)"
//
// The smallest earthbend card there is: the keyword and nothing else,
// which is why it is the catalog's proof that the verb works. Four
// mana for a 4/4 hasty attacker that also taps for mana, and whose
// downside is mostly notional — it comes back tapped when it dies, so
// a removal spell costs the opponent a card and costs you one untap.
//
// The Lesson subtype is type-line data only; there is no Learn in the
// catalog (see the Paradigm row in docs/engine-seams.md for the other
// Lesson-adjacent gap, which is a different keyword).
//
// EVERYTHING IS THE KEYWORD ACTION. The animation, the haste, the four
// counters and the delayed return all live in
// `game.EarthbendForEffect`; this file is the count and the shared
// target clause. That is deliberate — twelve cards print this reminder
// text, and a per-card composition would be twelve chances to get the
// layer, the duration or the delayed trigger's key wrong.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e113f0f-4469-4276-82e1-0a804f3de444",
		Name:         "Earthbending Lesson",
		Completeness: CompletenessFull,
		Targets:      EarthbendTargets(),
		OnResolve:    EarthbendFirstTarget(EarthbendCount(4)),
	})
}
