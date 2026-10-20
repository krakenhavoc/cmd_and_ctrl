package effects

// Yavimaya Hollow — Legendary Land:
//
//	"{T}: Add {C}.
//	 {G}, {T}: Regenerate target creature."
//
// A colorless mana ability and a regeneration ability on one land; the
// two share {T}, so tapping for mana spends the regeneration and the
// other way round. The second is an ordinary activated ability built
// like Asceticism's — {G} plus {T}, target creature, the shared
// regenerateTheTargetPermanent effect — so it uses the stack and can be
// answered, and the shield it makes replaces the NEXT destruction of
// that creature this turn (CR 701.19a). Legendary is printed type-line
// data the engine already carries.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53d6113d-acdb-4754-9641-f7991a96c7b9",
		Name:         "Yavimaya Hollow",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}: Regenerate target creature.",
			Cost:    Plus(ManaCost("{G}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  regenerateTheTargetPermanent,
		}},
	})
}
