package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atzocan Seer — Creature — Human Druid {1}{G}{W}, 2/3 (Lost Caverns of
// Ixalan Commander):
//
//	"{T}: Add one mana of any color.
//	 Sacrifice this creature: Return target Dinosaur card from your
//	 graveyard to your hand."
//
// The sacrifice is a cost (CR 118.3) paid before anything resolves, and
// the target is a Dinosaur card in the graveyard as the ability goes on the stack.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "affced7f-797d-4e50-8a24-2da14de44e43",
		Name:         "Atzocan Seer",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Return target Dinosaur card from your graveyard to your hand.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCardInGraveyard("target Dinosaur card from your graveyard", effects.YouOwn(), effects.Subtype("Dinosaur")),
			Effect:  effects.ReturnTargetedCardToHand,
		}},
	})
}
