package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wish — Sorcery {2}{R}:
//
//	"You may play a card you own from outside the game this turn."
//
// GrantOutsidePlayRight (#3955, ADR 0198's 2026-10-16 amendment): when
// Wish resolves it names no card. It stores a standing, one-use,
// until-end-of-turn permission over the owner's cards outside the game,
// and the card is chosen when the player plays it, at whatever timing
// the card allows, as printed. With a sideboard the owner casts a card
// from the sideboard tab; without one the `outside_play` action opens a
// pick over the card database, mints the named card and goes straight
// into its cast or land play. A land counts (it is the turn's land
// drop). Cancelling the pick spends nothing; a card played does.
func init() {
	Register(Spec{
		OracleID:     "2ccbf0ee-c856-4f67-8579-811fe8571a2e",
		Name:         "Wish",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return GrantOutsidePlayRight{Name: "Wish"}.Apply(ctx)
		},
	})
}
