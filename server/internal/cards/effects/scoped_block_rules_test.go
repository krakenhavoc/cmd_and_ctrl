package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scoped_block_rules_test.go pins ADR 0041 phase 3 tier 3b for the
// block rules (#1497): Gingerbrute, Departed Deckhand's granted
// evasion, and Mirri, Weatherlight Duelist's per-defender limit are
// ScopedEffect records, so a table holding one is a restore point.

// scopedBlockRuleKinds are the tier 3b block-rule kinds.
var scopedBlockRuleKinds = map[game.ModKind]bool{
	game.ModCantBeBlockedExceptBy:    true,
	game.ModLimitBlockersPerDefender: true,
}

// scopedBlockRuleCount is how many live records hold a block-rule
// mod — what `len(g.TurnScopedBlockRules)` used to answer.
func scopedBlockRuleCount(g *game.Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, e := range g.ScopedEffects {
			for _, m := range e.Mods {
				if scopedBlockRuleKinds[m.Kind] {
					n++
					break
				}
			}
		}
	})
	return n
}
