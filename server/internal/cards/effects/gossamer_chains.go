package effects

// Gossamer Chains — {W}{W} Enchantment:
//
//	"Return this enchantment to its owner's hand: Prevent all combat
//	 damage that would be dealt by target unblocked creature this turn."
//
// #2028: the return is the cost (ReturnThis), paid as the ability is
// activated (CR 602.2b), so the enchantment is in its owner's hand
// before anyone can respond. The shield is Kor Haven's (ADR 0108 §7,
// Delivery PR 7b): the targeted creature, pinned as the ability
// resolves, deals no combat damage to anything for the rest of the
// turn. "Unblocked" is the reading ninjutsu's cost uses
// (UnblockedAttacker).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "af3936ab-0156-438d-9fef-a924d2fcffc3",
		Name:         "Gossamer Chains",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{shieldAgainstTargetsRow(
			"Return this enchantment to its owner's hand: Prevent all combat damage that would be dealt by target unblocked creature this turn.",
			ReturnThis(),
			TargetCreature("target unblocked creature", UnblockedAttacker()), true)},
	})
}
