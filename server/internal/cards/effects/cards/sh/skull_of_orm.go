package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skull of Orm — Artifact {3} (Eighth Edition, S82 older sets batch 619, #5607):
//
//	"{5}, {T}: Return target enchantment card from your graveyard to your
//	 hand."
//
// Argivian Archaeologist's graveyard return with an enchantment card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67601399-089b-44fe-82db-c319729c75bf",
		Name:         "Skull of Orm",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Return target enchantment card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target enchantment card in your graveyard", effects.Enchantment(), effects.YouOwn()),
			Effect:  effects.ReturnTargetedCardToHand,
		}},
	})
}
