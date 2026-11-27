package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Changeling Titan — Creature — Shapeshifter {4}{G}, 7/7 (Lorwyn, S82 OLD b548,
// #5593):
//
//	"Changeling (This card is every creature type.)
//	 Champion a creature (When this enters, sacrifice it unless you exile
//	 another creature you control. When this leaves the battlefield, that card
//	 returns to the battlefield.)"
//
// Champion (CR 702.72a) is the shared champion-a-creature body
// (NewGapB138ExileACreatureOrSacrifice, Wormfang Drake's): as the enters
// trigger resolves its controller picks another creature they control to
// exile until this leaves the battlefield (ExileUntil, so the return is the
// engine's record of the exile), or declines or has none and sacrifices this.
// A champion that has already left by then exiles nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c1409a11-3780-40e4-abe6-df53d26b2dd6",
		Name:            "Changeling Titan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Triggered:       []game.TriggeredAbility{changelingTitanChampion.Trigger()},
	})
}

var changelingTitanChampion = effects.NewGapB138ExileACreatureOrSacrifice("Changeling Titan", "choose-cards/s82-old-b548-changeling-titan-champion")
