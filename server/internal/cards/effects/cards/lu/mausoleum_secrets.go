package cardslu

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mausoleum Secrets — Instant {1}{B} (Guilds of Ravnica, #5525):
//
//	"Undergrowth — Search your library for a black card with mana value less
//	 than or equal to the number of creature cards in your graveyard, reveal
//	 it, put it into your hand, then shuffle."
//
// The creature cards are counted as the spell resolves (CR 608.2h). The
// search is for a stated quality, so the player may find nothing and the
// library is shuffled either way (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30c4914e-4d9f-453d-bf5b-48ab8e0f2ba3",
		Name:         "Mausoleum Secrets",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.B11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.HasColorWord("B") && c.ManaValue() <= n },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Source:    ctx.Source(),
				Reason:    "Mausoleum Secrets — search for a black card with mana value " + strconv.Itoa(n) + " or less",
			}.Apply(ctx)
		},
	})
}
