package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Make Obsolete — Instant {2}{B} (Kaladesh, #5539):
//
//	"Creatures your opponents control get -1/-1 until end of turn."
//
// The Black Breath's -1/-1, locked to the creatures its opponents
// control as it resolves (CR 611.2c), declared as the one-sided sweep it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ce755f5-855f-4b88-aa10-22f4ee76b28a",
		Name:         "Make Obsolete",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		OnResolve: effects.ResolveSteps(effects.BoostUntilEOT{
			Match:     effects.And(effects.Creature(), effects.OpponentControls()),
			Power:     -1,
			Toughness: -1,
			Label:     "Make Obsolete — -1/-1 until end of turn",
		}),
	})
}
