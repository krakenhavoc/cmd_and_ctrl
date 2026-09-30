package effects

// Annihilating Glare — Sorcery {B}:
//
//	"As an additional cost to cast this spell, pay {4} or sacrifice an
//	 artifact or creature. Destroy target creature or planeswalker."
//
// The either/or cost (ADR 0100 §2), branches in printed order: branch 0
// is the {4}, which is exactly why the announcement's branch is required
// rather than defaulted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a75087aa-41bd-4e0f-a118-61b5209356f5",
		Name:         "Annihilating Glare",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			ManaAdditionalCost("{4}").Keyed("mana"),
			SacrificeCost("an artifact or creature", Or(Artifact(), Creature())).Keyed("sacrifice"),
		),
		Targets:   TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: destroyTheTargetPermanent,
	})
}
