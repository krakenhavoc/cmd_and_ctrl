package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Talas Warrior — Creature — Human Pirate Warrior {1}{U}{U}, 2/2 (Portal Second Age,
// #5625):
//
//	"This creature can't be blocked."
//
// Triton Shorestalker's evasion: a RestrictSelf static the
// declare-blockers check reads, so no creature can be declared as its
// blocker (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12beabf3-a84b-49c2-9462-8876a4ac74d4",
		Name:         "Talas Warrior",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
	})
}
