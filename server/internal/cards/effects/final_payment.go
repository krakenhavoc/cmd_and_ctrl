package effects

// Final Payment — Instant {W}{B}:
//
//	"As an additional cost to cast this spell, pay 5 life or sacrifice a
//	 creature or enchantment. Destroy target creature."
//
// The either/or cost (ADR 0100 §2). The life branch is refused at
// announce for a caster below 5 life (CR 119.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e47d57dc-2e69-4939-8aec-077595f2ae05",
		Name:         "Final Payment",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			PayLifeCost(5).Keyed("life"),
			SacrificeCost("a creature or enchantment", Or(Creature(), Enchantment())).Keyed("sacrifice"),
		),
		Targets:   TargetCreature("target creature"),
		OnResolve: destroyTheTargetPermanent,
	})
}
