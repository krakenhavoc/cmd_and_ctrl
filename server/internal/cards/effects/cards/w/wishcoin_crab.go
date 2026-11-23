package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wishcoin Crab — Creature — Crab {3}{U}, 2/5 (Guilds of Ravnica, #5525):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "abc519b8-82de-4559-be03-ddb1ddaab993",
		Name:         "Wishcoin Crab",
		Completeness: effects.CompletenessFull,
	})
}
