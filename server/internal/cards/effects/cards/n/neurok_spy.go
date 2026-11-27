package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neurok Spy — Creature — Human Rogue {2}{U}, 2/2 (Mirrodin, #5606):
//
//	"This creature can't be blocked as long as defending player controls an
//	 artifact."
//
// Scrapdiver Serpent's block rule on the creature itself: the defending player
// is the one it attacks, and the test is made as blockers are declared
// (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d78cb7d8-9cfc-4d16-86ea-d49fc5d65db0",
		Name:         "Neurok Spy",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnSelf(), effects.DefendingPlayerControlsAnArtifact),
		},
	})
}
