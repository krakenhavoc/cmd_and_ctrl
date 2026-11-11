package effects

// Accursed Duneyard — Land:
//
//	"{T}: Add {C}.
//	 {2}, {T}: Regenerate target Shade, Skeleton, Specter, Spirit, Vampire,
//	 Wraith, or Zombie."
//
// A colourless mana ability and a regeneration shield (CR 701.19a) for a
// creature of one of the seven types, any controller. The target's types are
// read through the layered view, so a changeling qualifies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48edc348-93f6-4dce-9cc4-7244d76b6f4a",
		Name:         "Accursed Duneyard",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Regenerate target Shade, Skeleton, Specter, Spirit, Vampire, Wraith, or Zombie.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target Shade, Skeleton, Specter, Spirit, Vampire, Wraith, or Zombie",
				Or(HasSubtype("Shade"), HasSubtype("Skeleton"), HasSubtype("Specter"), HasSubtype("Spirit"),
					HasSubtype("Vampire"), HasSubtype("Wraith"), HasSubtype("Zombie"))),
			Effect: regenerateTheTargetPermanent,
		}},
	})
}
