package effects

// Deadly Precision — Sorcery {B}:
//
//	"As an additional cost to cast this spell, pay {4} or sacrifice an
//	 artifact or creature. Destroy target creature."
//
// The either/or cost (ADR 0100 §2), branches in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d23ca3d2-0731-4df4-a3c6-69ddf0c58601",
		Name:         "Deadly Precision",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			ManaAdditionalCost("{4}").Keyed("mana"),
			SacrificeCost("an artifact or creature", Or(Artifact(), Creature())).Keyed("sacrifice"),
		),
		Targets:   TargetCreature("target creature"),
		OnResolve: destroyTheTargetPermanent,
	})
}
