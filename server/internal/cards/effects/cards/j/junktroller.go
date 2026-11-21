package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Junktroller — Artifact Creature — Golem {4}, 0/6 (Ravnica Remastered,
// #5948):
//
//	"Defender
//	 {T}: Put target card from a graveyard on the bottom of its owner's
//	 library."
//
// The activated ability is Chrome Companion's: any graveyard's card goes to
// the bottom of its owner's library, and a card that left the graveyard in
// response is an illegal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ef46fd7-82db-45d3-b02e-7312384577f3",
		Name:            "Junktroller",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.SocB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
