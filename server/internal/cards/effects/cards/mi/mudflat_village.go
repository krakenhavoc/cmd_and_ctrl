package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mudflat Village — Land (Bloomburrow, #4467):
//
//	"{T}: Add {C}.
//	 {T}: Add {B}. Spend this mana only to cast a creature spell.
//	 {1}{B}, {T}, Sacrifice this land: Return target Bat, Lizard, Rat,
//	 or Squirrel card from your graveyard to your hand."
//
// Bloomburrow's black creature-type land, built like Lilypad Village:
// the {B} carries the cast and creature restriction tags, and is a
// deliberate click since the auto-tapper leaves restricted abilities
// alone. The sacrifice ability is Buried Ruin's, with the target
// narrowed to a card with one of the four subtypes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aeeab1df-0b8b-4bc4-a5f9-aac413449bec",
		Name:         "Mudflat Village",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{B}",
				Label:    "Add {B} (creature spells only)",
				Restrictions: []string{
					effects.ManaRestrictCast,
					effects.ManaRestrictType("Creature"),
				},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{B}, {T}, Sacrifice this land: Return target Bat, Lizard, Rat, or Squirrel card from your graveyard to your hand.",
			Cost:  effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("target Bat, Lizard, Rat, or Squirrel card in your graveyard",
				effects.AnySubtype("Bat", "Lizard", "Rat", "Squirrel"), effects.YouOwn()),
			Effect: effects.ReturnTargetedCardToHand,
		}},
	})
}
