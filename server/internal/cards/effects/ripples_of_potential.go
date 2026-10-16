package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ripples of Potential — Instant {1}{U} (EDHREC rank 881, #3654):
//
//	"Proliferate, then choose any number of permanents you control that
//	 had a counter put on them this way. Those permanents phase out."
//
// The proliferate is the player's choice as always (CR 701.34a), through
// game.ProliferateChoosingThenForEffect (#3930), whose continuation runs
// once every counter has landed and is told what really received one. A
// permanent that was chosen but got no counter (one replaced away, or one
// that left first) is not offered. Then a choose-cards prompt over those
// that are still on the battlefield under the caster's control, any
// number of them including none, and the picks phase out (CR 702.26),
// with anything attached to them (CR 702.26g). A proliferate that
// "proliferate twice" doubled offers everything either pass gave a
// counter to.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45bcebd0-5e93-4433-b92f-b4c161d04bd6",
		Name:         "Ripples of Potential",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller, source := ctx.Controller(), ctx.Source()
			return ctx.Game.ProliferateChoosingThenForEffect(controller, source, func(g *game.Game, got game.Proliferated) error {
				return ripplesChoosePhaseOut(g, controller, source, got.Permanents)
			})
		},
	})
}

// ripplesChoosePhaseOut asks `controller` which of the permanents the
// proliferate gave a counter to, among those they still control, phase
// out. Asks nothing when there are none. Captures scalars only.
func ripplesChoosePhaseOut(g *game.Game, controller, source uuid.UUID, countered []uuid.UUID) error {
	var mine []uuid.UUID
	for _, id := range countered {
		if !onBattlefield(g, id) {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && c.Controller == controller {
			mine = append(mine, id)
		}
	}
	if len(mine) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source,
		Question: "Ripples of Potential — choose any number of permanents you control that got a counter to phase out",
		Cards:    mine,
		Min:      0,
		Max:      0, // "any number"
		Zone:     game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return g.PhaseOutForEffect(source, picked...)
		},
	})
	return nil
}
