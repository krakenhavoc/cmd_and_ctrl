package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b807_helpers.go — S82 set OLD, batch 807 (Mirage, keyword creatures).

// OldB807NonartifactNonDragon is the condition of Catacomb Dragon's "becomes
// blocked by a nonartifact, non-Dragon creature": the blocker is not an
// artifact and does not have the Dragon creature type.
func OldB807NonartifactNonDragon(_ *game.Game, c game.Card) bool {
	return !c.IsArtifact() && !c.HasSubtypeWord("Dragon")
}

// OldB807HalfRoundedDown is "half the creature's power, rounded down" (CR
// 107.1a): rounded toward negative infinity, so half of -3 is -2 and half of 5
// is 2.
func OldB807HalfRoundedDown(n int) int {
	if n >= 0 {
		return n / 2
	}
	return -((-n + 1) / 2)
}
