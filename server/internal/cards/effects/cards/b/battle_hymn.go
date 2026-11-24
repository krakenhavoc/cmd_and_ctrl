package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Battle Hymn — Instant {1}{R}:
//
//	"Add {R} for each creature you control."
//
// An instant that uses the stack, so it can be countered (CR 605.3a, unlike a
// mana ability). The count is read as it resolves, on the layered view, so a
// creature that died in response is not counted and an animated land is; a
// player with no creature adds nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b977117-741c-4157-8fcf-719e475a8a4c",
		Name:         "Battle Hymn",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := len(effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl())))
			return effects.AddMana{Amount: effects.ManaN(n, "R")}.Apply(ctx)
		},
	})
}
