package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jade-Cast Sentinel — Artifact Creature — Ape Snake {4}, 1/5 (Tarkir:
// Dragonstorm, #4463):
//
//	"Reach
//	 {2}, {T}: Put target card from a graveyard on the bottom of its
//	 owner's library."
//
// Chrome Companion's activated ability: a card from any graveyard goes to
// the bottom of its owner's library, and a card that left the graveyard
// in response is an illegal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "985192cd-ceb6-48ad-80b9-e10fc544260e",
		Name:            "Jade-Cast Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.SocB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
