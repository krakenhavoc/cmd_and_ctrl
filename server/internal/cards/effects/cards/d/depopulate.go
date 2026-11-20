package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "4a83c2fa-f2d2-4b86-8407-4269f127936d",
		Name:         "Depopulate",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: effects.CompletenessFull,
		// The destruction waits for the prompts the draws ask (a dredge offer; CR
		// 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				drawers := map[uuid.UUID]bool{}
				for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.Multicolored())) {
					drawers[c.Controller] = true
				}
				for _, p := range ctx.Game.Seats {
					if p == nil || !drawers[p.ID] {
						continue
					}
					if err := (effects.DrawCards{Player: p.ID, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.DestroyAllMatching{Match: effects.Creature()},
		),
	})
}
