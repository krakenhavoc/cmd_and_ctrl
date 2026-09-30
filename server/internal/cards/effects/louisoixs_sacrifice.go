package effects

// Louisoix's Sacrifice — Instant {U}:
//
//	"As an additional cost to cast this spell, sacrifice a legendary
//	 creature or pay {2}. Counter target activated ability, triggered
//	 ability, or noncreature spell."
//
// The either/or cost (ADR 0100 §2); the target clause is Tale's End's
// shape (#1211) with "noncreature" where Tale's End says "legendary".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42d8ab49-703a-4fd3-84d0-0a3ac1eafca8",
		Name:         "Louisoix's Sacrifice",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			SacrificeCost("a legendary creature", Legendary(), Creature()).Keyed("sacrifice"),
			ManaAdditionalCost("{2}").Keyed("mana"),
		),
		Targets: TargetSpellOrAbility("target activated ability, triggered ability, or noncreature spell",
			AnyStackItem(AnAbilityItem(), ASpellItem(Noncreature()))),
		OnResolve: counterTheTargetSpell,
	})
}
