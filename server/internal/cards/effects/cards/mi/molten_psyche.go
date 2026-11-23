package cardsmi

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "6f98c1eb-0f52-4d21-a266-e8d52f98c1b3",
		Name:         "Molten Psyche",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.WindsOfChangeShuffleFrom(ctx.Game, effects.TablePlayers(ctx), nil)
			}),
			effects.Instruction(moltenPsycheMetalcraft),
		),
	})
}

// moltenPsycheMetalcraft is the metalcraft clause.
func moltenPsycheMetalcraft(ctx *effects.Context) error {
	if !effects.YouControlThreeOrMoreArtifacts(ctx.Game, ctx.Controller(), uuid.Nil) {
		return nil
	}
	return effects.S82BnsB09DamageEachOpponentBy(ctx, func(opp uuid.UUID) int {
		return ctx.Game.TurnTallyFor(opp).CardsDrawn
	})
}
