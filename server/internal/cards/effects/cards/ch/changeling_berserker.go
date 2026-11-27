package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Changeling Berserker — Creature — Shapeshifter {3}{R}, 5/3 (Lorwyn, S82 OLD b548,
// #5593):
//
//	"Changeling (This card is every creature type.)
//	 Haste
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
		OracleID:        "6a3dee57-9912-4c51-872e-2ec47acdaf52",
		Name:            "Changeling Berserker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "haste"},
		Triggered:       []game.TriggeredAbility{changelingBerserkerChampion.Trigger()},
	})
}

var changelingBerserkerChampion = effects.NewGapB138ExileACreatureOrSacrifice("Changeling Berserker", "choose-cards/s82-old-b548-changeling-berserker-champion")
