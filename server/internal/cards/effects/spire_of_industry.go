package effects

// Spire of Industry — Land:
//
//	"{T}: Add {C}.
//	 {T}, Pay 1 life: Add one mana of any color. Activate only if you
//	 control an artifact."
//
// Two mana abilities. The second carries the life payment as a cost
// component (`ManaAbilityCost.Life`) and the "activate only if" as a
// `ManaAbility.Condition`, checked before anything is paid, so with no
// artifact the ability is not offered and no life is lost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55a519b4-61cb-448a-875b-4d6dbe00580f",
		Name:         "Spire of Industry",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			painlessColorless(),
			{
				Cost:      ManaAbilityCost{Tap: true, Life: 1},
				Produced:  "{W|U|B|R|G}",
				Label:     "{T}, Pay 1 life: Add one mana of any color (only if you control an artifact)",
				Condition: ControlsAtLeast(1, MatchArtifact),
			},
		},
	})
}
