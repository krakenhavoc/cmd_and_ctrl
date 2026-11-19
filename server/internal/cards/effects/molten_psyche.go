package effects

import (
	"github.com/google/uuid"
)

// Molten Psyche — Sorcery {1}{R}{R}:
//
//	"Each player shuffles the cards from their hand into their library,
//	 then draws that many cards.
//	 Metalcraft — If you control three or more artifacts, Molten Psyche
//	 deals damage to each opponent equal to the number of cards that player
//	 has drawn this turn."
//
// Winds of Change's shuffle-and-redraw, then (a second step, so it waits
// for any prompt a draw asked, CR 608.2c) the metalcraft damage: Metalcraft
// is read as that step runs, and each opponent is dealt the cards they have
// drawn this turn, this redraw and their draw step's included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f98c1eb-0f52-4d21-a266-e8d52f98c1b3",
		Name:         "Molten Psyche",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return windsOfChangeShuffleFrom(ctx.Game, tablePlayers(ctx), nil)
			}),
			Instruction(moltenPsycheMetalcraft),
		),
	})
}

// moltenPsycheMetalcraft is the metalcraft clause.
func moltenPsycheMetalcraft(ctx *Context) error {
	if !youControlThreeOrMoreArtifacts(ctx.Game, ctx.Controller(), uuid.Nil) {
		return nil
	}
	return s82BnsB09DamageEachOpponentBy(ctx, func(opp uuid.UUID) int {
		return ctx.Game.TurnTallyFor(opp).CardsDrawn
	})
}
