package effects

// Spark Harvest — Sorcery {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 pay {3}{B}. Destroy target creature or planeswalker."
//
// The either/or cost (ADR 0100 §2); the {3}{B} branch joins the total at
// CR 601.2f through the one pricer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25756e33-25e5-4976-8cce-a39fa2c4b007",
		Name:         "Spark Harvest",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			SacrificeCost("a creature", Creature()).Keyed("sacrifice"),
			ManaAdditionalCost("{3}{B}").Keyed("mana"),
		),
		Targets:   TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: destroyTheTargetPermanent,
	})
}
