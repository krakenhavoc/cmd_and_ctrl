package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keeper of the Cadence — Creature — Human Wizard {4}{U}, 2/5 (The Brothers'
// War, #5482):
//
//	"{3}: Put target artifact, instant, or sorcery card from a graveyard on
//	 the bottom of its owner's library."
//
// Chrome Companion's tuck, with the card kind narrowed: any graveyard's
// artifact, instant or sorcery card goes to the bottom of its owner's
// library, and one that left in response is an illegal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b3c5fc6-6c4a-4e23-9625-5954d1986160",
		Name:         "Keeper of the Cadence",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Put target artifact, instant, or sorcery card from a graveyard on the bottom of its owner's library.",
			Cost:    effects.ManaCost("{3}"),
			Targets: effects.TargetCardInGraveyard("target artifact, instant, or sorcery card from a graveyard", effects.Or(effects.Artifact(), effects.Instant(), effects.Sorcery())),
			Effect:  effects.SocB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
