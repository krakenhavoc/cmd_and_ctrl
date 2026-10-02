package effects

// Lava Coil — Sorcery {1}{R}:
//
//	"Lava Coil deals 4 damage to target creature. If that creature would
//	 die this turn, exile it instead."
//
// The replacement is the spell's, not the damage's (ADR 0108 §1): the
// target is marked whether or not the damage is dealt, so a creature
// that survives the 4 damage and dies later this turn is still exiled.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "fa71db44-5181-4c51-8b24-7fbedf36e3ca",
		Name:         "Lava Coil",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    damageFirstTargetExileIfItDies(fixedAmount(4)),
	})
}
