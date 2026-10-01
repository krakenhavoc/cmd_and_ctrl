package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terramorph — Sorcery {3}{G}:
//
//	"Search your library for a basic land card, put it onto the
//	 battlefield, then shuffle.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The land enters untapped: the card says nothing about tapped.
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "17a34f8d-a80f-4331-8be5-06cbb9d10d7b",
		Name:            "Terramorph",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: IsBasicLand,
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Terramorph — a basic land",
			}.Apply(ctx)
		},
	})
}
