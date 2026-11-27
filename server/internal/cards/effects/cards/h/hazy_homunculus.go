package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hazy Homunculus — Creature — Homunculus Illusion {1}{U}, 1/1 (Prophecy, S82
// OLD b671, #5617):
//
//	"This creature can't be blocked as long as defending player controls an
//	 untapped land."
//
// Scrapdiver Serpent's CantBeBlockedWhile over the defending player's board,
// read as blockers are declared: an untapped land they control, by effective
// type.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1af26f29-627c-48e5-a19d-deb43a11a846",
		Name:         "Hazy Homunculus",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnSelf(), effects.OldB671DefendingPlayerControlsAnUntappedLand),
		},
	})
}
