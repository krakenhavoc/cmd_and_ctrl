package effects

// Martyr's Cause — Enchantment {2}{W}:
//
//	"Sacrifice a creature: The next time a source of your choice would deal damage to any target this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// "Any target" is a target (CR 115.4): a player, a creature, a
// planeswalker or a battle, chosen as the ability is activated; the
// source is chosen as it resolves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "72d1c789-e79f-4f15-80d0-981d4120b085",
		Name:         "Martyr's Cause",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{nextDamageShieldRow(
			"Sacrifice a creature: The next time a source of your choice would deal damage to any target this turn, prevent that damage.",
			SacrificeACreature(), TargetAny(), PreventNextDamageFromChosenSource(ShieldTheTarget))},
	})
}
