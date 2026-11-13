package effects

// Quiet Purity — Instant — Arcane {W} (Champions of Kamigawa, #5603):
//
//	"Destroy target enchantment."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f16dc14-3d3f-4ffa-90bc-9abc67db75cf",
		Name:         "Quiet Purity",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
