package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Entish Restoration — Instant {2}{G} (EDHREC rank 449):
//
//	"Sacrifice a land. Search your library for up to two basic land
//	 cards, put them onto the battlefield tapped, then shuffle. If
//	 you control a creature with power 4 or greater, instead search
//	 your library for up to three basic land cards, put them onto the
//	 battlefield tapped, then shuffle."
//
// Instant-speed ramp that nets one land, or two with a big creature
// out. The ferocious branch is decided as the spell resolves, using
// the same power-4 check Garruk's Uprising uses (CurrentPower, so
// counters and pumps count), and the search is one SearchLibrary
// with the limit set by that branch — "up to" is the chooser's
// minimum of zero.
//
// Sandbox simplification, the Victimize posture: "Sacrifice a land"
// is modelled as an ADDITIONAL COST TO CAST, not a resolution-time
// action, because no sacrifice-then-continue prompt exists. Every
// observable difference runs the weaker way — the land is gone even
// if the spell is countered, and the spell cannot be cast with no
// land to sacrifice (printed, it can be cast to search for nothing).
func init() {
	Register(Spec{
		OracleID:       "736017e2-bc33-49e8-812d-1639443fdb51",
		Name:           "Entish Restoration",
		AdditionalCost: SacrificeCost("a land", Land()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			limit := 2
			if youControlPowerFourOrGreater(ctx.Game, ctx.Controller()) {
				limit = 3
			}
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         limit,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Entish Restoration — basic lands, onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
