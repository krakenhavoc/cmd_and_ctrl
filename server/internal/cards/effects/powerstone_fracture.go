package effects

// Powerstone Fracture — Sorcery {1}{B} (The Brothers' War, S82):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature.
//	 Destroy target creature or planeswalker."
//
// Bone Splinters' shape: the sacrifice is paid while casting (CR 601.2h),
// after the target is chosen, so what is sacrificed cannot usefully be the
// target. With nothing to sacrifice the cast is refused.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "f13baf38-27fc-4c5d-8fee-459eefdb32da",
		Name:           "Powerstone Fracture",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		Targets:        TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve:      destroyTheTargetPermanent,
	})
}
