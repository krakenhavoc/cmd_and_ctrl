package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sneaky Homunculus — Creature — Homunculus Illusion {1}{U}, 1/1
// (Nemesis, #5618):
//
//	"This creature can't block or be blocked by creatures with power 2 or greater."
//
// CantBlockOrBeBlockedBy: two block rules, one per side of the pair (CR
// 509.1b), against creatures whose power is 2 or greater as blockers
// are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b76bb3b-1ada-4c78-bd13-3464371d4fa0",
		Name:         "Sneaky Homunculus",
		Completeness: effects.CompletenessFull,
		BlockRules:   effects.CantBlockOrBeBlockedBy(effects.OnSelf(), effects.PowerGE(2), "creatures with power 2 or greater"),
	})
}
