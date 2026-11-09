package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tend the Sprigs — Sorcery {2}{G} (Lorwyn Eclipsed, #4458):
//
//	"Search your library for a basic land card, put it onto the battlefield
//	 tapped, then shuffle. Then if you control seven or more lands and/or
//	 Treefolk, create a 3/4 green Treefolk creature token with reach."
//
// "Lands and/or Treefolk" counts each permanent once: a Dryad Arbor is one.
// The count is taken after the fetched land has entered, in the search's
// keyed tail, so it runs whether or not a prompt was needed.
func init() {
	Register(Spec{
		OracleID:     "91363d71-eef0-48e8-9a68-8f363ff2790e",
		Name:         "Tend the Sprigs",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				TappedOnEntry: true,
				Limit:         1,
				Shuffle:       true,
				Reason:        "Tend the Sprigs — a basic land card",
				Source:        item.SourceCardID,
				ThenTail:      ctx.ItemTail(tendTheSprigsTail),
			}.Apply(ctx)
		},
	})
}

// tendTheSprigsTail is the "then if you control seven or more lands and/or
// Treefolk" clause. An on-disk key: never renamed, never reused.
var tendTheSprigsTail = CardTail("search-then/tend-the-sprigs", func(ctx *Context, _ game.TailOutcome) error {
	treefolk := ctx.TextSubtype("Treefolk")
	n := countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool {
		return c.IsLand() || c.HasSubtype(treefolk)
	})
	if n < 7 {
		return nil
	}
	return CreateToken{Template: TokenCard("3/4 green Treefolk with reach"), N: 1}.Apply(ctx)
})
