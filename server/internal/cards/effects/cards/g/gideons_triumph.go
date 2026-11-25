package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gideon's Triumph — Instant {1}{W} (War of the Spark, #5520):
//
//	"Target opponent sacrifices a creature of their choice that attacked or
//	 blocked this turn. If you control a Gideon planeswalker, that player
//	 sacrifices two of those creatures of their choice instead."
//
// "Attacked or blocked this turn" is the turn tally's attack and block
// records, read per object (CR 400.7). The Gideon clause is read as the spell
// resolves (CR 608.2); with only one such creature the player sacrifices that
// one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "daec03b6-e3f8-4e44-9766-c64dc23fa3dc",
		Name:         "Gideon's Triumph",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var who uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					who = t.ID
				}
			}
			if who == uuid.Nil {
				return nil
			}
			asks := []uuid.UUID{who}
			if effects.B42ControlsPlaneswalkerNamed(ctx.Game, ctx.Controller(), "Gideon") {
				asks = append(asks, who)
			}
			ctx.Game.PlayersSacrificeForEffect(ctx.Source(), asks,
				effects.SacrificeSpec("a creature that attacked or blocked this turn", effects.S82OldB170AttackedOrBlocked()),
				"Gideon's Triumph — sacrifice a creature that attacked or blocked this turn")
			return nil
		},
	})
}
