package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Changeling Hero — Creature — Shapeshifter {4}{W}, 4/4 (Lorwyn, S82 OLD b548,
// #5593):
//
//	"Changeling (This card is every creature type.)
//	 Lifelink
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
		OracleID:        "fc893d4a-debd-4df0-ae83-d3ebd2f0fefe",
		Name:            "Changeling Hero",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "lifelink"},
		Triggered:       []game.TriggeredAbility{changelingHeroChampion.Trigger()},
	})
}

var changelingHeroChampion = effects.NewGapB138ExileACreatureOrSacrifice("Changeling Hero", "choose-cards/s82-old-b548-changeling-hero-champion")
