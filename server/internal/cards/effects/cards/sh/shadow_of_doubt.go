package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow of Doubt — Instant {U/B}{U/B} (Secret Lair Drop):
//
//	"({U/B} can be paid with either {U} or {B}.)
//	 Players can't search libraries this turn.
//	 Draw a card."
//
// From resolution until the turn ends no library can be searched
// (TurnTally.LibrarySearchesBarred, read in the one search primitive): a search
// of a library alone does not happen at all, with no search event, though the
// "then shuffle" after it is its own instruction and still happens; the
// graveyard or hand half of an and/or search is unaffected. The draw follows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4b724bb-4c1c-4002-8eb8-bc8dd3f88e62",
		Name:         "Shadow of Doubt",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				ctx.Game.BarLibrarySearchesThisTurnForEffect()
				return nil
			}),
			effects.DrawCards{N: 1},
		),
	})
}
