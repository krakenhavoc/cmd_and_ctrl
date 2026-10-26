package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Faithless Salvaging — Instant {1}{R}:
//
//	"Discard a card, then draw a card.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// A rummage: the discard is chosen first, and the draw waits for it
// (a registered continuation, #4253). With an empty hand nothing is discarded and the
// card is still drawn (CR 701.9a). Rebound is the engine's keyword
// (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0366ccfd-c717-4ea2-8176-86e184f920f4",
		Name:            "Faithless Salvaging",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// "Discard a card, then draw a card", keyed (#4253): the draw
			// is Charming Scoundrel's registered rummage.
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player: ctx.Controller(),
				Source: item.SourceCardID,
				N:      1,
			}, charmingScoundrelDraw, nil)
		},
	})
}
