package effects

// Honored Heirloom — Artifact {3} (Innistrad: Crimson Vow):
//
//	"{T}: Add one mana of any color.
//	 {2}, {T}: Exile target card from a graveyard."
//
// A mana rock with a graveyard-hate ability. Both abilities tap it, so
// they compete for the one tap. The exile targets any card in any
// graveyard and is re-checked on resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bf0427e-5e87-4d32-ad34-1d8946ab96fd",
		Name:         "Honored Heirloom",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Exile target card from a graveyard.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Purpose: ForTargets(ExilesFromGraveyard(0, "")),
			Effect:  b27ExileChosenTarget,
		}},
	})
}
