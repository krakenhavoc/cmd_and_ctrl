package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flux — Sorcery {2}{U} (Weatherlight, S82 older sets batch 777, #5629):
//
//	"Each player discards any number of cards, then draws that many cards.
//	 Draw a card."
//
// Each player, from the active player in turn order, picks how many cards to discard from the hand
// they hold (any number, none included) and then draws that many (the run's per-seat tail); the
// caster's own draw follows once every player has answered (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8cff5f1-6871-49d2-89a7-24870308aadb",
		Name:         "Flux",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB777EachPlayerDiscardsAnyNumberThenDraws(ctx, "Flux — discard any number of cards, then draw that many", fluxRest)
		},
	})
}

// fluxRest is the per-seat "then draws that many cards" and the run's "Draw a card" for the
// caster. An on-disk key.
var fluxRest = game.RegisterPromptRunThenLeg("prompt-run/s82-old-b777-flux", effects.S82OldB777DrawAsManyAsDiscarded,
	func(g *game.Game, r game.RunDone) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return g.DrawNForEffect(r.Carry[0], 1)
	})
