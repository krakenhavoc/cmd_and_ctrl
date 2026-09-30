package effects

// Bitter Triumph — Instant {1}{B}:
//
//	"As an additional cost to cast this spell, discard a card or pay 3
//	 life. Destroy target creature or planeswalker."
//
// The either/or cost (ADR 0100 §2). The "pay 3 life" branch is refused
// at announce for a caster below 3 life (CR 119.4), and the view shows
// it disabled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "776341cb-d2ec-423f-9250-92dc8bd8d503",
		Name:         "Bitter Triumph",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			DiscardCost(1).Keyed("discard"),
			PayLifeCost(3).Keyed("life"),
		),
		Targets:   TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: destroyTheTargetPermanent,
	})
}
