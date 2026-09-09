package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyshroud Claim — Sorcery {3}{G}:
//
//	"Search your library for up to two Forest cards, put them onto
//	the battlefield, then shuffle."
//
// Two lands, UNTAPPED, and "Forest card" rather than basic — so it
// fetches duals. Net-positive on mana the turn it resolves, which is
// why it beats Explosive Vegetation in decks that can support it.
//
// Two sequential single-card searches rather than one Limit: 2 call:
// the sandbox picker takes the first predicate match, so two passes
// with the shuffle deferred to the second is the same idiom Cultivate
// uses to get two distinct cards in library order.
func init() {
	Register(Spec{
		OracleID: "376c9d3f-21d3-4251-bb6a-026fa9e1b0e1",
		Name:     "Skyshroud Claim",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			if err := (SearchLibrary{
				Player:    controller,
				Predicate: IsLandWithSubtype("forest"),
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Reveal:    true,
				Shuffle:   false,
			}).Apply(ctx); err != nil {
				return err
			}
			return SearchLibrary{
				Player:    controller,
				Predicate: IsLandWithSubtype("forest"),
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
			}.Apply(ctx)
		},
	})
}
