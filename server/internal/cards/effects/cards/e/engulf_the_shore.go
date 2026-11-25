package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Engulf the Shore — Instant {3}{U} (Game Night 2019, #5515):
//
//	"Return to their owners' hands all creatures with toughness less than or
//	 equal to the number of Islands you control."
//
// Both numbers are read as it resolves: the Islands the caster controls (a
// land that is an Island by type, basic or not), then every creature, whoever
// controls it, whose toughness (after every effect) is no more than that,
// all returned as one simultaneous move. With no Island nothing has a
// toughness that low unless it is 0 or less, and those have already died.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3d4ed7c-c586-4581-96dc-977636d27583",
		Name:         "Engulf the Shore",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			islands := len(effects.MatchingBattlefield(ctx, effects.And(effects.HasSubtype("Island"), effects.YouControl())))
			return effects.BounceAllMatching{Match: effects.And(effects.Creature(), effects.ToughnessLE(islands))}.Apply(ctx)
		},
	})
}
