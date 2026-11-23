package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fading Hope — Instant {U} (Innistrad: Midnight Hunt, #5497):
//
//	"Return target creature to its owner's hand. If its mana value was 3 or
//	 less, scry 1."
//
// The mana value is read before the creature leaves (CR 608.2h, last-known
// information); the scry is the bounce's continuation, so it follows the
// card reaching a hand or the command zone question being answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64883ef2-c536-42e2-b52f-cfd1984fcdfd",
		Name:         "Fading Hope",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Scry: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := effects.BattlefieldCardForEffect(ctx.Game, id)
			if !ok {
				return nil
			}
			b := effects.BounceToHand{Target: id}
			if mv, known := ctx.Game.ManaValueForEffect(c); known && mv <= 3 {
				b.ThenTail = ctx.TailN(effects.S82MidThenScry, 1)
			}
			return b.Apply(ctx)
		},
	})
}
