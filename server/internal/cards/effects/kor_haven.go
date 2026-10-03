package effects

// Kor Haven — Legendary Land:
//
//	"{T}: Add {C}.
//	 {1}{W}, {T}: Prevent all combat damage that would be dealt by target attacking creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted attacker, pinned as the ability resolves (CR 400.7); its
// combat damage to anything is prevented for the rest of the turn.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "276cece9-f9f2-46e6-ae76-daddaa2fb9ab",
		Name:         "Kor Haven",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{shieldAgainstTargetsRow(
			"{1}{W}, {T}: Prevent all combat damage that would be dealt by target attacking creature this turn.",
			Plus(ManaCost("{1}{W}"), TapCost()),
			TargetCreature("target attacking creature", AttackingCreature()), true)},
	})
}
