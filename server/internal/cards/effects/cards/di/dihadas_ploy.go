package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dihada's Ploy — Instant {1}{U}{B} (Modern Horizons 2, #5501):
//
//	"Draw two cards, then discard a card. You gain life equal to the number
//	 of cards you've discarded this turn.
//	 Jump-start"
//
// The two cards are drawn, then the discard is chosen; the life is read once
// the discard has happened, off the turn's discard tally (CR 701.9a), so the
// discard just made and any earlier this turn all count. With an empty hand
// nothing is discarded and the tally is read as it stands. Jump-start is
// the shared graveyard offer that also discards a card as a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "cc929343-db65-43d5-9bf6-9f0bd77c2cd8",
		Name:             "Dihada's Ploy",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Draws: 2, LifeGain: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Cmdr2S39JumpStart("{1}{U}{B}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player: ctx.Controller(),
				Source: item.SourceCardID,
				N:      1,
			}, effects.OldB032GainLifePerDiscardThisTurn, nil)
		},
	})
}
