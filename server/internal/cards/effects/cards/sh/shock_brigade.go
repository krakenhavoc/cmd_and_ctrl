package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shock Brigade — Creature — Goblin Soldier {1}{R}, 1/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 Mobilize 1 (Whenever this creature attacks, create a tapped and
//	 attacking 1/1 red Warrior creature token. Sacrifice it at the
//	 beginning of the next end step.)"
//
// Mobilize is mobilize.go (CR 702.181a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8676e917-17dd-4047-99c4-402f2e909e16",
		Name:            "Shock Brigade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered:       []game.TriggeredAbility{effects.Mobilize("Shock Brigade", 1)},
	})
}
