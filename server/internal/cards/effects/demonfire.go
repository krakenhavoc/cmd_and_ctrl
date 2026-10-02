package effects

// Demonfire — Sorcery {X}{R}:
//
//	"Demonfire deals X damage to any target. If a creature dealt damage
//	 this way would die this turn, exile it instead.
//	 Hellbent — If you have no cards in hand, this spell can't be
//	 countered and the damage can't be prevented."
//
// "A creature dealt damage this way" is registered from the damage's
// continuation, on a creature that was dealt more than 0 damage (ADR
// 0108 §1 decision 3). Both hellbent riders are the spell's own (ADR
// 0107 §5), judged when something tries to counter it and as it deals
// its damage, so a Demonfire cast with an empty hand goes through a Fog
// and a shield.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                   "314a5c76-1a68-433b-a383-1834400254a8",
		Name:                       "Demonfire",
		XMatters:                   true,
		Completeness:               CompletenessFull,
		CantBeCounteredIf:          SpellHellbent(),
		SpellDamageCantBePrevented: SpellHellbent(),
		Targets:                    TargetAny(),
		OnResolve:                  damageAnyTargetExileIfDealtDies(xAmount(), false),
	})
}
