package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sun's Bounty — Instant {1}{W} (Coldsnap):
//
//	"You gain 4 life.
//	 Recover {1}{W}"
//
// Recover (CR 702.59a) is S82OldB564Recover.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c15d9b1-5ec7-4476-a3e6-539d899a9957",
		Name:         "Sun's Bounty",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.GainLife{Amount: 4}),
		Triggered:    []game.TriggeredAbility{effects.S82OldB564Recover("Sun's Bounty", "{1}{W}")},
	})
}
