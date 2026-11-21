package effects

// Abstruse Archaic — Creature — Avatar {4}, 3/4 (Commander Masters):
//
//	"Vigilance
//	 {1}, {T}: Copy target activated or triggered ability you control from
//	 a colorless source. You may choose new targets for the copy. (Mana
//	 abilities can't be targeted.)"
//
// Lithoform Engine's copy (copyTargetedAbility: the copy is its controller's,
// with new targets on offer, CR 707.10c) over an ability of the caster's
// from a colorless source. The source's colour is read as the ability is
// chosen, from the permanent as it is now or as it last was (a sacrificed
// colorless artifact's ability still qualifies), by
// gapB98AbilityFromColorlessSource. Mana abilities are not on the stack and
// so are never offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0727dc95-03df-4a2c-aed5-73fb840f3db7",
		Name:            "Abstruse Archaic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Copy target activated or triggered ability you control from a colorless source. You may choose new targets for the copy.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: AbilityOnStack("target activated or triggered ability you control from a colorless source", gapB98AbilityFromColorlessSource()),
			Effect:  copyTargetedAbility,
		}},
	})
}
