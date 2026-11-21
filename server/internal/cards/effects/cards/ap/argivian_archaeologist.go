package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Argivian Archaeologist — Creature — Human Artificer {1}{W}{W}, 1/1
// (Antiquities, #5640):
//
//	"{W}{W}, {T}: Return target artifact card from your graveyard to
//	 your hand."
//
// Buried Ruin's recursion on a creature, with a mana and a tap cost in
// place of the land's sacrifice: the same graveyard clause and the same
// resolution body (returnTargetedCardToHand). The tap is a {T} cost, so
// the Archaeologist cannot use it the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b889ec0-6130-4e31-bb02-03fdabd28bee",
		Name:         "Argivian Archaeologist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{W}, {T}: Return target artifact card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{W}{W}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target artifact card in your graveyard", effects.Artifact(), effects.YouOwn()),
			Effect:  effects.ReturnTargetedCardToHand,
		}},
	})
}
