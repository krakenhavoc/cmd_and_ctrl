package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renari, Merchant of Marvels — Legendary Creature — Dragon Artificer
// {3}{U}, 2/4 (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"You may cast Dragon spells and artifact spells as though they had flash.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// Shimmer Myr's cast-timing rule with the permission filter's "artifact or
// <subtype>" shape (Arcade Gannon's), whose subtype is read through the
// card's creature types. Choose a Background is a deck-construction rule
// (CR 702.124k) that internal/deck reads off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99204279-2a50-4dd6-84ec-f84ba31554cf",
		Name:         "Renari, Merchant of Marvels",
		Completeness: effects.CompletenessFull,
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{ArtifactOrSubtype: "Dragon"},
				"You may cast Dragon spells and artifact spells as though they had flash."),
		},
	})
}
