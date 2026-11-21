package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peer into the Abyss — Sorcery {4}{B}{B}{B} (EDHREC rank 1272):
//
//	"Target player draws cards equal to half the number of cards in
//	 their library and loses half their life. Round up each time."
//
// Half a library and half a life total, both rounded up, both read
// at resolution. The draw comes first, as printed; the life loss is
// loss, not damage, so no prevention applies. A life total already at
// or below zero loses nothing (CR 107.1b — a negative loss is zero).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21fa2442-6eac-4dce-a9cc-76f0053fdb8f",
		Name:         "Peer into the Abyss",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.ResolveSteps(
			// The life loss halves the life after the draw, and waits for a prompt
			// the draw asks (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				p := peerIntoTheAbyssTarget(ctx)
				if p == nil {
					return nil
				}
				draw := 0
				if p.Library != nil {
					draw = (p.Library.Size() + 1) / 2
				}
				return effects.DrawCards{Player: p.ID, N: draw}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				p := peerIntoTheAbyssTarget(ctx)
				if p == nil {
					return nil
				}
				lose := (p.Life + 1) / 2
				if lose <= 0 {
					return nil
				}
				return effects.GainLife{Player: p.ID, Amount: -lose}.Apply(ctx)
			}),
		),
	})
}

// peerIntoTheAbyssTarget is the targeted player while they are still in
// the game, or nil. Each of the spell's steps reads it from the game, since
// a piece cannot hand it to the next (#4155).
func peerIntoTheAbyssTarget(ctx *effects.Context) *game.Player {
	if len(ctx.Item.Targets) == 0 || ctx.Item.Targets[0].Kind != game.TargetPlayer {
		return nil
	}
	p := ctx.PlayerByID(ctx.Item.Targets[0].ID)
	if p == nil || p.Eliminated {
		return nil
	}
	return p
}
