package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mishra's Juggernaut — Artifact Creature — Juggernaut {5}, 5/3:
//
//	"Trample
//	 This creature attacks each combat if able.
//	 Unearth {5}{R}"
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fa6fa2f5-eb9b-402e-82f4-e0a8d30df245",
		Name:            "Mishra's Juggernaut",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
		Activated:       []effects.ActivatedAbility{effects.Unearth("{5}{R}")},
	})
}
