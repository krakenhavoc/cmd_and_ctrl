package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scrapdiver Serpent — Creature — Serpent {5}{U}{U}, 5/5 (Commander Legends):
//
//	"This creature can't be blocked as long as defending player controls an
//	 artifact."
//
// A block rule on the Serpent itself (CantBeBlockedWhile): the defending player
// is the one it attacks, or the controller of the planeswalker it attacks, and
// the test is made as blockers are declared (CR 509.1b), so an artifact that
// arrives after blockers does not unblock it and one destroyed before them
// makes it blockable.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f45ac24-e153-4ec1-a41f-989f9547d409",
		Name:         "Scrapdiver Serpent",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnSelf(), effects.DefendingPlayerControlsAnArtifact),
		},
	})
}
