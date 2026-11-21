package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "7e103748-3f76-42ce-a063-d0256b2dce2b",
		Name:         "Rockface Village",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{R}",
				Label:    "Add {R} (creature spells only)",
				Restrictions: []string{
					effects.ManaRestrictCast,
					effects.ManaRestrictType("Creature"),
				},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{R}, {T}: Target Lizard, Mouse, Otter, or Raccoon you control gets +1/+0 and gains haste until end of turn. Activate only as a sorcery.",
			Cost:  effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target Lizard, Mouse, Otter, or Raccoon you control",
				effects.AnySubtype("Lizard", "Mouse", "Otter", "Raccoon"), effects.YouControl()),
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "haste")),
			SorcerySpeed: true,
			Effect:       effects.PumpAndGrantFirstTarget(1, 0, "Rockface Village — +1/+0 and haste", "haste"),
		}},
	})
}
