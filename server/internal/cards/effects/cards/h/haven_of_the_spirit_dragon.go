package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haven of the Spirit Dragon — Land:
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a Dragon
//	 creature spell.
//	 {2}, {T}, Sacrifice this land: Return target Dragon creature card or
//	 Ugin planeswalker card from your graveyard to your hand."
//
// Beastcaller Savant's restricted any-colour mana, restricted to Dragon
// creature spells (Carnelian Orb's restriction triple); the mana pays any
// part of such a spell's cost, additional or alternative included. The
// return is Memorial to Folly's, aimed at either kind of card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acc9c16a-5e72-43bd-87e1-56a16aa892f5",
		Name:         "Haven of the Spirit Dragon",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictType("Creature"), effects.ManaRestrictSubtype("Dragon")},
				Label:        "Add one mana of any color. Spend this mana only to cast a Dragon creature spell",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice this land: Return target Dragon creature card or Ugin planeswalker card from your graveyard to your hand.",
			Cost:  effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("target Dragon creature card or Ugin planeswalker card in your graveyard",
				effects.YouOwn(), effects.Or(effects.And(effects.Creature(), effects.Subtype("Dragon")), effects.And(effects.Planeswalker(), effects.Subtype("Ugin")))),
			Effect: effects.B27ReturnChosenGraveyardCardToHand,
		}},
	})
}
