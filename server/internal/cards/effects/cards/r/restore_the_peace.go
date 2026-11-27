package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restore the Peace — Instant {1}{W}{U} (Dragon's Maze, S82 OLD batch 403, #5564):
//
//	"Return each creature that dealt damage this turn to its owner's hand."
//
// Red Guardian's reading of "dealt damage this turn" (a damage event this turn
// naming the creature as its source, an amount above zero, wiped if the creature
// entered after it, CR 400.7) over every creature on the battlefield as the spell
// resolves, bounced together as one event (Cyclonic Rift's mass bounce); tokens
// cease to exist. Nothing is targeted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aeca0661-8b71-4085-81e9-bb0884fc9831",
		Name:         "Restore the Peace",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		OnResolve: effects.ResolveSteps(effects.BounceAllMatching{
			Match: effects.And(effects.Creature(), effects.MshA6DealtDamageThisTurn()),
		}),
	})
}
