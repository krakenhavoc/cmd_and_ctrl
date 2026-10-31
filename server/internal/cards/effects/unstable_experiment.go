package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unstable Experiment — Instant {1}{U} (Marvel Super Heroes Commander,
// #4455):
//
//	"Target player draws a card, then up to one target creature you control
//	 connives."
//
// Two target clauses (Arc Trail's Clauses): slot 0 the player, slot 1 an
// optional creature you control. Each is re-checked as the spell resolves
// (CR 608.2b), so a player who became illegal draws nothing and the
// creature still connives, and the other way round. The creature connives
// for its controller as the spell resolves (CR 701.50a, draw then discard
// then a counter for a nonland card), through connive.go's ConniveEach.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b7bbc76-22cc-482e-8d65-34cff34c93c8",
		Name:         "Unstable Experiment",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPlayer("target player"),
			UpToOneTargetCreature("up to one target creature you control", YouControl()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if p, ok := ctx.ClauseTarget(0); ok && p.Kind == game.TargetPlayer {
				if err := (DrawCards{Player: p.ID, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			c, ok := ctx.ClauseTarget(1)
			if !ok || c.Kind != game.TargetCard {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(c.ID)
			if !ok {
				return nil
			}
			return ConniveEach{Controller: card.Controller, Creatures: []uuid.UUID{c.ID}}.Apply(ctx)
		},
	})
}
