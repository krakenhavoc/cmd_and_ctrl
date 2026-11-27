package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dream Prowler — Creature — Illusion {2}{U}{U}, 1/5 (Stronghold, S82 OLD b750,
// #5627):
//
//	"This creature can't be blocked as long as it's attacking alone."
//
// Bubbling Beebles' conditional unblockability (CantBeBlockedWhile over itself)
// with CR 506.5's "attacking alone", read live when the block is checked: the
// Prowler is attacking and no other creature is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b56c5b39-2465-4a25-b806-bb84866e576f",
		Name:         "Dream Prowler",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, you uuid.UUID, c game.Card) bool {
			return effects.AttackingAlone()(g, you, c)
		})},
	})
}
