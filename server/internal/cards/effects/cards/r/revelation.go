package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Revelation — World Enchantment {G}:
//
//	"Players play with their hands revealed."
//
// #3953, ADR 0199, one of #1997's world enchantments: every player's hand,
// its controller's too, is shown to every player while Revelation is on
// the battlefield (game.HandRevealedLocked). A world permanent: the world
// rule (CR 704.5k, ADR 0109 §8) puts it into its owner's graveyard when a
// newer one enters.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "251ff4e4-3be4-424e-8dd7-eb4ede7b415c",
		Name:          "Revelation",
		Completeness:  effects.CompletenessFull,
		HandsRevealed: game.HandRevealEachPlayer,
	})
}
