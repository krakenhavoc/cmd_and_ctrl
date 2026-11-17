package effects

// Lava Spike — Sorcery — Arcane {R} (Champions of Kamigawa, #5603):
//
//	"Lava Spike deals 3 damage to target player or planeswalker."
//
// Skullcrack's target clause without its riders. An Arcane spell, so it
// can carry a splice (CR 702.47).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2837888c-bfa8-4955-9334-0605ad409f7e",
		Name:         "Lava Spike",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		Targets:      targetPlayerOrPlaneswalker(),
		OnResolve:    damageToFirstTarget(3),
	})
}
