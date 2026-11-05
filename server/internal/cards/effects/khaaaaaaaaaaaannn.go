package effects

// Khaaaaaaaaaaaannn! — Instant {X}{R} (Star Trek, #4973):
//
//	"Khaaaaaaaaaaaannn! deals twice X damage to target creature. If that
//	 creature would die this turn, exile it instead."
//
// "That creature" is the target, not "a creature dealt damage this way",
// so the replacement is the spell's own effect and is set up on a legal
// target whether or not the damage is dealt (Scorching Dragonfire's shape,
// ADR 0108 §1). The damage is twice the announced X.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45b1a19b-6a0c-4534-9106-99c2e5fc30d6",
		Name:         "Khaaaaaaaaaaaannn!",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    damageFirstTargetExileIfItDies(twiceXAmount()),
	})
}
