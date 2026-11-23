package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cower in Fear — Instant {1}{B}{B} (Magic 2013, #5568):
//
//	"Creatures your opponents control get -1/-1 until end of turn."
//
// A one-sided layer 7c shrink (CR 613.4c) on the creatures your
// opponents control as it resolves (CR 611.2c), ending at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1519fe3-af13-4491-9020-4a8d8f474713",
		Name:         "Cower in Fear",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -1, Toughness: -1,
				Label: "Cower in Fear — -1/-1 until end of turn"}.Apply(ctx)
		},
	})
}
