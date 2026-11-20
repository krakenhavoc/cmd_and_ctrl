package effects

// Spinning Wheel — Artifact {3} (Throne of Eldraine, #5516):
//
//	"{T}: Add one mana of any color.
//	 {5}, {T}: Tap target creature."
//
// A Mana Geode's mana ability and the shared tap row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0935a0f-7087-40cd-a094-89f200da6ef0",
		Name:         "Spinning Wheel",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			samiAnyColorMana(ManaAbilityCost{Tap: true}, "{T}: Add one mana of any color"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
