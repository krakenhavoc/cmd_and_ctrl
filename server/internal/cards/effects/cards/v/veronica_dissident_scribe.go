package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veronica, Dissident Scribe — Legendary Creature — Human Artificer Rogue
// {2}{R}, 3/3 (Fallout, #5948):
//
//	"Menace
//	 Whenever Veronica attacks, you may discard a card. If you do, draw a
//	 card.
//	 Whenever you discard one or more nonland cards for the first time each
//	 turn, create a Junk token. (It's an artifact with "{T}, Sacrifice this
//	 token: Exile the top card of your library. You may play that card this
//	 turn. Activate only as a sorcery.")"
//
// The attack trigger is Ashling's rummage (b39MayDiscardThenDraw): the discard
// first, a draw only for a card really discarded. The discard trigger is one
// per batch of discards (ThatManyEvents, Conspiracy Theorist's shape), counting
// only nonland cards, and only the first such batch of the turn
// (TriggersOnlyOnceEachTurn: a land discarded first leaves it unspent, and a
// second nonland discard the same turn makes nothing). The Junk is the
// printed token, and Veronica's own attack rummage counts.
//
// No simplification.
func init() {
	const discardLabel = "Veronica, Dissident Scribe — create a Junk token"
	effects.Register(effects.Spec{
		OracleID:        "5bf01e86-d5f2-4639-9a07-f442fd1c3bb4",
		Name:            "Veronica, Dissident Scribe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Veronica — you may discard a card; if you do, draw a card", func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, "Veronica, Dissident Scribe — discard a card, then draw a card",
					effects.DrawAsManyAsDiscarded)(effects.NewContext(g, item))
			}),
			effects.WithAtBatchEnd(effects.TriggersOnlyOnceEachTurn(effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.DiscardedByYou(ev, source) && effects.ConspiracyTheoristNonland(g, ev.CardID)
			}, discardLabel, effects.Do(effects.CreateToken{Template: effects.JunkToken(), N: 1}))), effects.ThatManyEvents()),
		},
	})
}
