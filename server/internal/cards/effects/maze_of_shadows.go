package effects

// Maze of Shadows — Land:
//
//	"{T}: Add {C}.
//	 {T}: Untap target attacking creature with shadow. Prevent all combat
//	 damage that would be dealt to and dealt by that creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): Maze of Ith's ability, narrowed to
// a creature with shadow, beside a colourless mana ability.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "b7b51ab1-403e-4640-8827-b04965aa6760",
		Name:         "Maze of Shadows",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{untapAttackerToAndByRow(
			"{T}: Untap target attacking creature with shadow. Prevent all combat damage that would be dealt to and dealt by that creature this turn.",
			TapCost(), TargetCreature("target attacking creature with shadow", AttackingCreature(), HasKeyword("shadow")))},
	})
}
