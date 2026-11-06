package effects

// Vote Out — Sorcery {3}{B} (Edge of Eternities, #4461):
//
//	"Convoke
//	 Destroy target creature."
//
// Pile On's convoke (TapCost: Convoke(), CR 702.51a) on a plain
// destroy: each creature tapped while casting it pays for {1} or one
// mana of that creature's color.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7efe9fe5-d93c-46a6-afbd-9655a902c6bc",
		Name:         "Vote Out",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      TargetCreature("target creature"),
		OnResolve:    destroyTheTargetPermanent,
	})
}
