package effects

// Bayou Groff — Creature — Plant Dog {1}{G}, 5/4:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 pay {3}."
//
// The either/or cost (ADR 0100 §2) and nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e83b2790-3cec-429c-a65d-6f4c7d025d37",
		Name:         "Bayou Groff",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			SacrificeCost("a creature", Creature()).Keyed("sacrifice"),
			ManaAdditionalCost("{3}").Keyed("mana"),
		),
	})
}
