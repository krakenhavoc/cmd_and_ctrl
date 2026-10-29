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
// game.ProliferateChoosingThenKeyedForEffect (#3930, #4253), whose keyed
// continuation runs once every counter has landed and is told what really
// received one, so a table waiting on any of its prompts is a restore
// point. A
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
			return ctx.Game.ProliferateChoosingThenKeyedForEffect(ctx.Controller(), ctx.Source(), ctx.Tail(ripplesProliferated))
		},
	})
}

// ripplesProliferated is the rest of the sentence once the proliferate has
// settled, told what got a counter: the permanents are the first Amount
// IDs. An on-disk key.
var ripplesProliferated = game.RegisterTail("then/ripples-of-potential-proliferated", func(g *game.Game, o game.TailOutcome) error {
	n := min(max(o.Amount, 0), len(o.IDs))
	return ripplesChoosePhaseOut(g, o.Tail.Player, o.Tail.Source, o.IDs[:n])
})

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
		LiveZone: game.ZoneBattlefield,
		ThenKey:  ripplesPhaseOut,
	})
	return nil
}

// ripplesPhaseOut is the permanents chosen phasing out, keyed (#4253): the
// card that asked is the prompt's source.
var ripplesPhaseOut = game.RegisterChooseCardsThen("choose-cards/ripples-of-potential-phase-out", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.PhaseOutForEffect(r.Source, r.Picked...)
})
