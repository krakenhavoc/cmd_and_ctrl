package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Watcher in the Web — Creature — Spider {4}{G}, 2/5:
//
//	"Reach
//	 This creature can block an additional seven creatures each combat."
//
// Reach is the engine's keyword; the second line is
// CanBlockAdditional(7) on itself (#1706) — up to eight attackers, with
// its 2 damage divided among them (CR 510.1d).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "232a4128-133e-446d-9f1c-dd3875094473",
		Name:            "Watcher in the Web",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static:          []game.StaticAbility{effects.CanBlockAdditional(effects.SelfOnly, 7)},
	})
}
