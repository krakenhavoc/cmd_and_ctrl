package effects

// Life's Legacy — Sorcery {1}{G}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw cards equal to the sacrificed creature's power."
//
// The sacrificed creature's power as it last existed on the battlefield
// (CR 608.2h), read off the payment record (ADR 0113 §1). A creature
// with zero or negative power draws nothing (CR 107.1b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "cb7baa46-7963-4d45-9a6e-e3c34db3c0e6",
		Name:           "Life's Legacy",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve:      drawEqualToSacrificedPower,
	})
}
