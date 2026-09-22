package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inspiring Call — Instant {2}{G} (EDHREC rank 277):
//
//	"Draw a card for each creature you control with a +1/+1 counter
//	 on it. Those creatures gain indestructible until end of turn."
//
// Both halves read the same set — creatures the controller controls
// carrying a +1/+1 counter — so one predicate serves the draw count
// and GrantKeywordUntilEOT's affected-set snapshot (CR 611.2c) alike.
// The until-end-of-turn machinery is #279's, unblocked for the whole
// batch.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b9a10ff-5a5d-4df8-88aa-18d84ff9117c",
		Name:         "Inspiring Call",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g := ctx.Game
			n := 0
			for _, c := range g.BattlefieldCardsForEffect() {
				if inspiringCallHasCounteredCreature(g, item.Controller, c) {
					n++
				}
			}
			if err := (DrawCards{Player: item.Controller, N: n}).Apply(ctx); err != nil {
				return err
			}
			if n == 0 {
				return nil
			}
			return GrantKeywordUntilEOT{
				Match: func(g *game.Game, _ uuid.UUID, c game.Card) bool {
					return inspiringCallHasCounteredCreature(g, item.Controller, c)
				},
				Keywords: []string{"indestructible"},
				Label:    "Inspiring Call — indestructible until end of turn",
			}.Apply(ctx)
		},
	})
}

// inspiringCallHasCounteredCreature is "a creature you control with a
// +1/+1 counter on it".
func inspiringCallHasCounteredCreature(_ *game.Game, controller uuid.UUID, c game.Card) bool {
	return c.Controller == controller && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0
}
