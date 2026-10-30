package effects

// Flame Slash — Sorcery {R}:
//
//	"Flame Slash deals 4 damage to target creature."
//
// The creature-only Lightning Bolt (Lightning Axe's and Concentrated Fire's
// body): one target clause that only a creature fits, so a player, a
// planeswalker and a battle are not offered, and one DealDamage from the
// spell. The target is judged at announce (CR 601.2c) and again as the
// spell resolves (CR 608.2b), so a creature that left or gained hexproof in
// response takes nothing and the spell does not resolve.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d98d674-6811-4d45-b22a-63792e272a2b",
		Name:         "Flame Slash",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		OnResolve:    damageToFirstTarget(4),
	})
}
