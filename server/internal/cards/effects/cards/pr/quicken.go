package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quicken — Instant {U}:
//
//	"The next sorcery spell you cast this turn can be cast as though it
//	 had flash. (It can be cast any time you could cast an instant.)
//	 Draw a card."
//
// The flash promise (#1852) is spent by the first sorcery spell cast
// this turn, at whatever speed it is cast; an instant or creature spell
// leaves it. The draw is unconditional and follows the promise.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf188dd3-1927-481b-b8c0-93b1222dbf53",
		Name:         "Quicken",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GrantNextSpellPromise{From: "Quicken", Promise: game.NextSpellPromise{
				Filter: game.PermissionFilter{SorceryOnly: true},
				Flash:  true,
				Text:   "The next sorcery spell you cast this turn can be cast as though it had flash.",
			}}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
