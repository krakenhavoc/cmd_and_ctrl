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
// S22: ONE search with Limit 2 rather than the two single-card
// passes this used to need. The old shape existed only because the
// picker took the first match in library order, so a second pass was
// the only way to reach a second card; with a real chooser the
// player is simply shown every Forest card and takes up to two —
// which is what "up to two" meant all along, and which is where the
// card's decision (basic Forest or Forest-typed dual?) actually
// lives.
func init() {
	Register(Spec{
		OracleID: "376c9d3f-21d3-4251-bb6a-026fa9e1b0e1",
		Name:     "Skyshroud Claim",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: IsLandWithSubtype("forest"),
				Dest:      game.ZoneBattlefield,
				Limit:     2,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Skyshroud Claim — up to two Forest cards",
			}.Apply(ctx)
		},
	})
}
