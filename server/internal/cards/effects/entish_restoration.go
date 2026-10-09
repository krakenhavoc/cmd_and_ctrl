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
// "Sacrifice a land" is part of the EFFECT, not a cost (#2876, the
// Victimize pattern of #2863, ADR 0013's amendment of 2026-10-09). It
// is a one-seat sacrifice run (PlayerSacrificesThenForEffect): the
// caster picks a land as the spell resolves and the search follows,
// whether or not a land was sacrificed ("Sacrifice a land. Search..." is
// not an "if you do"). So a countered spell costs no land, and the
// spell can be cast with no land at all. The power-4 check is read
// after the sacrifice, as the printed sequence has it.
func init() {
	Register(Spec{
		OracleID:     "736017e2-bc33-49e8-812d-1639443fdb51",
		Name:         "Entish Restoration",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return ctx.Game.PlayerSacrificesThenForEffect(
				ctx.Source(), controller,
				sacrificeSpec("a land", Land()),
				"Entish Restoration — sacrifice a land",
				1,
				func(g *game.Game, _ game.PromptedSacrifices) error {
					ctx := NewContext(g, item)
					limit := 2
					if youControlPowerFourOrGreater(g, controller) {
						limit = 3
					}
					return SearchLibrary{
						Player:        controller,
						Predicate:     IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         limit,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Entish Restoration — basic lands, onto the battlefield tapped",
					}.Apply(ctx)
				})
		},
	})
}
