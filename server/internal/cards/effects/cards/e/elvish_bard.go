package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Bard — Creature — Elf Shaman Bard, {3}{G}{G}, 2/4:
//
//	"All creatures able to block this creature do so."
//
// Prized Unicorn's Lure (#1684): the Bard's own ability, so a Bard
// that loses all abilities stops luring (CR 613.1f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "094b778f-95a0-436f-a9d0-20a46a674486",
		Name:         "Elvish Bard",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AllAbleToBlockDoSo()},
	})
}
