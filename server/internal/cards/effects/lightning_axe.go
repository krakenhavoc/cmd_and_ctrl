package effects

// Lightning Axe — Instant {R}:
//
//	"As an additional cost to cast this spell, discard a card or pay
//	 {5}. Lightning Axe deals 5 damage to target creature."
//
// The either/or cost (ADR 0100 §2). The "pay {5}" branch joins the total
// at CR 601.2f through the one pricer, so the auto-tapper, the preview
// and the bot all charge {5}{R} for it, and a cost modifier sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81b90905-fbc0-426a-a084-c3300533abb4",
		Name:         "Lightning Axe",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			DiscardCost(1).Keyed("discard"),
			ManaAdditionalCost("{5}").Keyed("mana"),
		),
		Targets:   TargetCreature("target creature"),
		Purpose:   ForTargets(DamageToTarget(0, 5)),
		OnResolve: damageToFirstTarget(5),
	})
}
