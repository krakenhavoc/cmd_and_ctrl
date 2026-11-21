package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thraben Heretic — Creature — Human Wizard {1}{W}, 2/2 (Dark Ascension, #5571):
//
//	"{T}: Exile target creature card from a graveyard."
//
// Any player's graveyard. A card that left the graveyard in response is
// left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67247945-17b7-4cb8-ba65-d0dfec367b77",
		Name:         "Thraben Heretic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Exile target creature card from a graveyard.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCardInGraveyard("target creature card in a graveyard", effects.Creature()),
			Effect:  effects.ExileFirstLegalCardTarget,
		}},
	})
}
