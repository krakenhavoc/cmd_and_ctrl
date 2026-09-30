package effects

// Stir Up Trouble — Sorcery {B}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature or pay {4}. Destroy target creature."
//
// The either/or cost (ADR 0100 §2), branches in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dda607bd-f419-4b7f-b052-a5ce6ce22bfe",
		Name:         "Stir Up Trouble",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			SacrificeCost("an artifact or creature", Or(Artifact(), Creature())).Keyed("sacrifice"),
			ManaAdditionalCost("{4}").Keyed("mana"),
		),
		Targets:   TargetCreature("target creature"),
		OnResolve: destroyTheTargetPermanent,
	})
}
