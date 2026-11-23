package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wormfang Drake — Creature — Nightmare Drake {2}{U}, 3/4 (Dominaria
// Remastered, #5948):
//
//	"Flying
//	 When this creature enters, sacrifice it unless you exile a creature you
//	 control other than this creature.
//	 When this creature leaves the battlefield, return the exiled card to
//	 the battlefield under its owner's control."
//
// The linked pair is GapB138ExileACreatureOrSacrifice: Champion's shape
// with "a creature" as the quality.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d37c4e77-5023-4d13-9cfe-a43f513364bf",
		Name:            "Wormfang Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			wormfangDrakeExile.Trigger(),
		},
	})
}

// wormfangDrakeExile is the enters trigger and its pick. The key is an
// on-disk identity.
var wormfangDrakeExile = effects.NewGapB138ExileACreatureOrSacrifice("Wormfang Drake", "choose-cards/s82-gap-b138-wormfang-drake-exile")
