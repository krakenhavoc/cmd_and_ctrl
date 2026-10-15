package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Green Sun's Zenith — Sorcery {X}{G} (EDHREC top 1,000, #3654):
//
//	"Search your library for a green creature card with mana value X or
//	 less, put it onto the battlefield, then shuffle. Shuffle Green Sun's
//	 Zenith into its owner's library."
//
// Chord of Calling's search for a green creature, then Blue Sun's
// Zenith's tuck. The search's own shuffle is left off: the tuck shuffles
// the owner's library with the spell in it, which is the only shuffle
// whose result can be seen. The spell goes back whether or not the search
// found a card (the two sentences are independent), and it goes back only
// if it is still on the stack (CR 608.2n), which the tuck checks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d96b60b-a060-48ee-bb83-93f1c4a10669",
		Name:         "Green Sun's Zenith",
		Completeness: CompletenessFull,
		XMatters:     true,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			src := ctx.Source()
			owner := item.Owner
			return SearchLibrary{
				Player: ctx.Controller(),
				Predicate: func(c game.Card) bool {
					return c.IsCreature() && c.ManaValue() <= x && slices.Contains(c.Colors, "G")
				},
				Dest:   game.ZoneBattlefield,
				Limit:  1,
				Reason: "Green Sun's Zenith — a green creature card with mana value X or less",
				Source: src,
				Then: func(g *game.Game, _ []uuid.UUID) error {
					return g.TuckToLibraryThenForEffect(src, game.TuckOptions{}, func(g *game.Game, _ bool) error {
						return g.ShuffleLibraryForEffect(owner)
					})
				},
			}.Apply(ctx)
		},
	})
}
