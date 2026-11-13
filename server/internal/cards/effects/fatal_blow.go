package effects

// Fatal Blow — Instant {B} (Weatherlight, #5629):
//
//	"Destroy target creature that was dealt damage this turn. It can't be
//	 regenerated."
//
// Qutrub Forayer's target clause. The destruction ignores a
// regeneration shield (CR 701.19c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16774611-c004-4383-b4e1-4a43a9b1f4aa",
		Name:         "Fatal Blow",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature that was dealt damage this turn", DealtDamageThisTurn()),
		OnResolve:    destroyTheTargetPermanentNoRegen,
	})
}
