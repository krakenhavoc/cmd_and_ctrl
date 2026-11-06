package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plunge into Winter — Instant {1}{W} (Wilds of Eldraine, #4471):
//
//	"Tap up to one target creature. Scry 1, then draw a card."
//
// The tap is Elder Deep-Fiend's over "up to one" target, so the spell may
// be cast with none. Then Opt: the draw is the scry's registered tail, so
// the card drawn is the one the player chose to leave on top. A spell cast
// with a target that has become illegal does not resolve at all
// (CR 608.2b); one cast with no target always scries and draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a272f15d-caa2-40e9-894f-19638bd5f375",
		Name:         "Plunge into Winter",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Scry: 1},
		Targets:      TargetCreature("up to one target creature").WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := tapEachLegalTarget(ctx.Game, item); err != nil {
				return err
			}
			c := ctx.Controller()
			return Scry{Player: c, N: 1, ThenTail: game.KeyedTail{Ref: thenDrawACard, Player: c}}.Apply(ctx)
		},
	})
}
