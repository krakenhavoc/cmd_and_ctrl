package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tainted Indulgence — Instant {U}{B} (Streets of New Capenna):
//
//	"Draw two cards. Then discard a card unless there are five or more
//	 mana values among cards in your graveyard."
//
// The draw lands first; the graveyard is then counted (the two draws change
// no graveyard) and the discard is the player's own choice from a hand that
// includes the drawn cards, asked only below five mana values
// (taintedIndulgenceDiscard, the Shoreline Looter shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6944092f-22c6-4b65-840d-8c1b70883127",
		Name:         "Tainted Indulgence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Discards: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 2, game.DrawThen{
				Ref: taintedIndulgenceDiscard, Player: item.Controller, Source: item.SourceCardID, N: 1,
			})
		},
	})
}

// taintedIndulgenceDiscard is "Then discard a card unless there are five or
// more mana values among cards in your graveyard". An on-disk key: never
// renamed, never reused.
var taintedIndulgenceDiscard = game.RegisterDrawThen("tainted-indulgence/discard-unless-five-mana-values", func(g *game.Game, d game.DrawThen) error {
	if effects.SncFiveMVs(g, d.Player) {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: d.Player, Source: d.Source, N: d.N})
	return nil
})
