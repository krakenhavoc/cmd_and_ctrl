package effects

// Axebane Guardian — Creature — Human Druid {2}{G}, 0/3 (Tarkir:
// Dragonstorm Commander):
//
//	"Defender
//	 {T}: Add X mana in any combination of colors, where X is the number
//	 of creatures you control with defender."
//
// A mana ability (it never uses the stack) in Selvala's shape: X
// independent any-colour picks, counted as the ability is activated and
// the Guardian included. The count reads defender through the layers, so
// a creature that lost it or was granted it counts as it stands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "060e0378-e335-4b3d-82aa-ae36639339f6",
		Name:            "Axebane Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedAnyCombinationOfColors(cmdrB19DefenderProduced),
			Label:          "{T}: Add X mana in any combination of colors, where X is the number of creatures you control with defender",
		}},
	})
}
