package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruin in Their Wake — Sorcery {1}{G}, devoid (Jumpstart 2022, #5948):
//
//	"Devoid (This card has no color.)
//	 Search your library for a basic land card and reveal it. You may put
//	 that card onto the battlefield tapped if you control a land named
//	 Wastes. Otherwise, put that card into your hand. Then shuffle."
//
// With a land named Wastes under your control as the spell resolves, the
// find may go onto the battlefield tapped and declining leaves it in the
// library (the "otherwise" is the no-Wastes case); without one it goes to
// your hand. Either way the card is revealed and the library shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68adbc52-34dd-47c4-bb1c-0c51033fe7cc",
		Name:            "Ruin in Their Wake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Answers: game.AnswerValue, Lands: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			wastes := len(effects.MatchingBattlefield(ctx, effects.And(effects.Land(), effects.YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return game.CardNameMatches(c, "Wastes") }))) > 0
			search := effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsBasicLand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Ruin in Their Wake — a basic land card",
			}
			if wastes {
				search.Dest = game.ZoneBattlefield
				search.TappedOnEntry = true
				search.Optional = true
			} else {
				search.Dest = game.ZoneHand
			}
			return search.Apply(ctx)
		},
	})
}
