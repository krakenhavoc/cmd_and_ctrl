package effects

// Rockface Village — Land (Bloomburrow, #4467):
//
//	"{T}: Add {C}.
//	 {T}: Add {R}. Spend this mana only to cast a creature spell.
//	 {R}, {T}: Target Lizard, Mouse, Otter, or Raccoon you control gets
//	 +1/+0 and gains haste until end of turn. Activate only as a
//	 sorcery."
//
// Bloomburrow's red creature-type land, built like Mudflat Village: the
// {R} carries the cast and creature restriction tags, and is a
// deliberate click since the auto-tapper leaves restricted abilities
// alone. The pump is a sorcery-speed activated ability whose target is a
// permanent you control with one of the four subtypes, re-checked as it
// resolves (pumpAndGrantFirstTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e103748-3f76-42ce-a063-d0256b2dce2b",
		Name:         "Rockface Village",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{R}",
				Label:    "Add {R} (creature spells only)",
				Restrictions: []string{
					ManaRestrictCast,
					ManaRestrictType("Creature"),
				},
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{R}, {T}: Target Lizard, Mouse, Otter, or Raccoon you control gets +1/+0 and gains haste until end of turn. Activate only as a sorcery.",
			Cost:  Plus(ManaCost("{R}"), TapCost()),
			Targets: TargetPermanent("target Lizard, Mouse, Otter, or Raccoon you control",
				AnySubtype("Lizard", "Mouse", "Otter", "Raccoon"), YouControl()),
			Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "haste")),
			SorcerySpeed: true,
			Effect:       pumpAndGrantFirstTarget(1, 0, "Rockface Village — +1/+0 and haste", "haste"),
		}},
	})
}
