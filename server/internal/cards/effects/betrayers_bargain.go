package effects

// Betrayer's Bargain — Instant {1}{R}:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 enchantment or pay {2}.
//	 Betrayer's Bargain deals 5 damage to target creature. If that
//	 creature would die this turn, exile it instead."
//
// The either/or cost (ADR 0100 §2), branches in printed order. The
// replacement is the spell's (ADR 0108 §1): the target is marked whether
// or not the damage is dealt.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "bfb3d862-d92d-4a4f-9a22-671b55d954fd",
		Name:         "Betrayer's Bargain",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			SacrificeCost("a creature or enchantment", Or(Creature(), Enchantment())).Keyed("sacrifice"),
			ManaAdditionalCost("{2}").Keyed("mana"),
		),
		Targets:   TargetCreature("target creature"),
		Purpose:   ForTargets(DamageToTarget(0, 5)),
		OnResolve: damageFirstTargetExileIfItDies(fixedAmount(5)),
	})
}
