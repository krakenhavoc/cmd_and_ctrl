package effects

// Hour of Defeat — Instant {3}{B} (Marvel Super Heroes, #4455):
//
//	"Destroy target creature. Surveil 1."
//
// Surveil happens whether or not the creature was destroyed (an
// indestructible target survives and the player still surveils). If the
// one target is illegal on resolution the spell does not resolve at all
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2774fcb7-8988-4a18-b5ec-7534dd8b2e05",
		Name:         "Hour of Defeat",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    destroyTargetThenSurveil(1),
	})
}
