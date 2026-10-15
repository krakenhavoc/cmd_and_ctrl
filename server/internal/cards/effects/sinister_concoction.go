package effects

// Sinister Concoction — Enchantment {B}:
//
//	"{B}, Pay 1 life, Mill a card, Discard a card, Sacrifice this
//	 enchantment: Destroy target creature."
//
// Five cost components on one ability, all paid at announce
// (CR 602.2b): the mana, the life, the mill (MillCardsCost, #3845, ADR
// 0020's 2026-10-15 amendment), a discard picked by the activator, and
// the sacrifice. The engine validates every component before paying
// any of them, so an empty library or an empty hand leaves the
// enchantment where it is (CR 118.3), and pays the mill after the
// others (CR 601.2h). The enchantment is gone by the time the ability
// resolves, which changes nothing: the destroy reads only the target.
//
// A target that has left, or stopped being a creature, by resolution
// is skipped (CR 608.2b); a cost already paid is not refunded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ffc3beec-6485-428d-a251-fa0bc21263a5",
		Name:         "Sinister Concoction",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}, Pay 1 life, Mill a card, Discard a card, Sacrifice this enchantment: Destroy target creature.",
			Cost:    Plus(ManaCost("{B}"), PayLife(1), MillCardsCost(1), DiscardACard(), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
