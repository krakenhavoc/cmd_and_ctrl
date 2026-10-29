package effects

// Stern Scolding — Instant {U} (The Hobbit Commander):
//
//	"Counter target creature spell with power or toughness 2 or less."
//
// #4454, S76. Essence Scatter's clause with a power-or-toughness cut:
// enforced at announce and again at resolution, so a spell whose
// numbers were changed in between is no longer a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3bc8a406-3ae7-4f56-a33c-b4a2c115b917",
		Name:         "Stern Scolding",
		Completeness: CompletenessFull,
		Targets: TargetSpell("target creature spell with power or toughness 2 or less",
			Creature(), hobB3PowerOrToughnessAtMost(2)),
		OnResolve: counterTheTargetSpell,
	})
}
