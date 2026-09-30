package effects

// Wild Unraveling — Instant {U}{U}:
//
//	"As an additional cost to cast this spell, blight 2 or pay {1}. (To
//	 blight 2, put two -1/-1 counters on a creature you control.) Counter
//	 target spell."
//
// The either/or cost (ADR 0100 §2), with the #1703 blight component in a
// branch.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd1a4219-437b-494f-ae0e-1d3181ffcd63",
		Name:         "Wild Unraveling",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			BlightCost(2).Keyed("blight"),
			ManaAdditionalCost("{1}").Keyed("mana"),
		),
		Targets:   TargetSpell("target spell"),
		OnResolve: counterTheTargetSpell,
	})
}
