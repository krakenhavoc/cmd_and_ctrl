package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aclazotz_attack.go — "each opponent discards a card; for each opponent who
// can't, you draw a card", shared by Aclazotz, Deepest Betrayal and
// Refurbished Familiar.

// AclazotzAttack makes each opponent discard a card of their choice, as one
// run, then draws a card for each opponent who didn't (aclazotzDraw).
func AclazotzAttack(g *game.Game, item *game.StackItem) error {
	return g.EachPlayerDiscardsThenKeyedForEffect(item.Controller,
		game.DiscardPrompt{Source: item.SourceCardID, N: 1},
		aclazotzDraw, []uuid.UUID{item.Controller})
}

// aclazotzDraw is "for each opponent who can't, you draw a card", run once
// after the last opponent's discard has landed: an opponent who discarded
// no card, because they had none to discard, is one. The controller is the
// carry. An on-disk key: never renamed, never reused.
var aclazotzDraw = game.RegisterPromptRunThen("prompt-run/aclazotz-draw", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	you := r.Carry[0]
	discards := r.Discards()
	cant := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you {
			continue
		}
		if len(discards.By(p.ID)) == 0 {
			cant++
		}
	}
	if cant == 0 {
		return nil
	}
	return g.DrawNForEffect(you, cant)
})
