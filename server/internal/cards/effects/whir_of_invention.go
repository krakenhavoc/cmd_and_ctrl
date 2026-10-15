package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whir of Invention — Instant {X}{U}{U}{U}: "Improvise ... Search your
// library for an artifact card with mana value X or less, put it onto
// the battlefield, then shuffle."
//
// Improvise is Spec.TapCost. The search is a stated-quality search, so it
// may find nothing. No simplification.
func init() {
	Register(Spec{
		OracleID:     "152b91c9-cc07-4ca8-944f-9bc2242a2283",
		Name:         "Whir of Invention",
		Completeness: CompletenessFull,
		XMatters:     true,
		TapCost:      Improvise(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			return SearchLibrary{
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
