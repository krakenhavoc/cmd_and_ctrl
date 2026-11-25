package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Krark's Thumb — {2} Legendary Artifact (Champions of Kamigawa, S82 OLD
// b028, #6152):
//
//	"If you would flip a coin, instead flip two coins and ignore one."
//
// A CoinFlipRule (game/coin_flip_rules.go), derived off the battlefield
// each time its controller flips: every coin becomes two, and the player
// keeps one. On a called flip they keep a won coin when either is won
// (CR 705.2); on an uncalled one, the face the effect wants for them. The
// ignored coin emits nothing, so "whenever you win a coin flip" sees one
// flip. Two Thumbs flip four coins and keep one. The oracle_id is the one
// the dump keeps on the reversible printing's face.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a97c8482-775c-4f11-9872-25b4f9bfcb1a",
		Name:         "Krark's Thumb",
		Completeness: effects.CompletenessFull,
		CoinFlipRules: []game.CoinFlipRule{
			{Label: "Krark's Thumb", FlipTwoIgnoreOne: true},
		},
	})
}
