package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Depopulate — Sorcery {2}{W}{W}:
//
//	"Each player who controls a multicolored creature draws a card.
//	 Then destroy all creatures."
//
// A four-mana Wrath of God that pays a card to anyone holding gold
// creatures — usually including you, since a deck playing a
// four-mana wrath in Commander is rarely mono-coloured.
//
// # Order matters and is printed
//
// "Then" is not filler. The draws happen FIRST, while the creatures
// are still on the battlefield, so who draws is decided by the board
// before the wipe. Destroying first and then checking would mean
// nobody ever draws. The two clauses are two statements here for
// exactly that reason.
//
// Seat order for the draws rather than map order, so a replay
// reproduces its own event log.
func init() {
	Register(Spec{
		OracleID:     "4a83c2fa-f2d2-4b86-8407-4269f127936d",
		Name:         "Depopulate",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Indestructible saves a permanent from single-target removal and from lethal damage, but a board wipe (\"destroy all\") still destroys it."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			drawers := map[uuid.UUID]bool{}
			for _, c := range MatchingBattlefield(ctx, And(Creature(), Multicolored())) {
				drawers[c.Controller] = true
			}
			for _, p := range ctx.Game.Seats {
				if p == nil || !drawers[p.ID] {
					continue
				}
				if err := (DrawCards{Player: p.ID, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return DestroyAllMatching{Match: Creature()}.Apply(ctx)
		},
	})
}
