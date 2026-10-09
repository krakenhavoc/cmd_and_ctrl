package effects

// Silence the Echo — Sorcery {1}{B} (Reality Fracture, tracker #2795):
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 planeswalker or pay {3}. Destroy target creature or planeswalker."
//
// The either/or additional cost (ADR 0100 §2), Annihilating Glare's shape,
// branches in printed order: the sacrifice, then the {3}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43654381-c55c-4d6d-abfd-12932ceb5a4a",
		Name:         "Silence the Echo",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			SacrificeCost("a creature or planeswalker", Or(Creature(), Planeswalker())).Keyed("sacrifice"),
			ManaAdditionalCost("{3}").Keyed("mana"),
		),
		Targets:   TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: destroyTheTargetPermanent,
	})
}
