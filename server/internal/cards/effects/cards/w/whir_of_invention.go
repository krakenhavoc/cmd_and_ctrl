package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whir of Invention — Instant {X}{U}{U}{U}: "Improvise ... Search your
// library for an artifact card with mana value X or less, put it onto
// the battlefield, then shuffle."
//
// Improvise is Spec.TapCost. The search is a stated-quality search, so it
// may find nothing. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "152b91c9-cc07-4ca8-944f-9bc2242a2283",
		Name:         "Whir of Invention",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		TapCost:      effects.Improvise(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.IsArtifact() && c.ManaValue() <= x },
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Whir of Invention — artifact card with mana value X or less",
				Source:    item.SourceCardID,
			}.Apply(ctx)
		},
	})
}
