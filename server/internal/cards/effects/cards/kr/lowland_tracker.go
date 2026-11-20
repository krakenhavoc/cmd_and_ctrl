package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lowland Tracker — Creature — Human Soldier, {4}{W}, 2/2:
//
//	"First strike
//	 Provoke (Whenever this creature attacks, you may have target
//	 creature defending player controls untap and block it if able.)"
//
// #1684 leftover: first strike rides PrintedKeywords, Provoke rides
// the Provoke() trigger constructor built for Goblin Grappler. No new
// machinery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "de06b33f-0793-419a-bf65-1961634af8e7",
		Name:            "Lowland Tracker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered:       []game.TriggeredAbility{effects.Provoke()},
	})
}
