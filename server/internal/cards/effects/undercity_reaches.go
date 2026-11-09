package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undercity Reaches — Plane — Ravnica (#4870):
//
//	"Whenever a creature deals combat damage to a player, its controller may
//	 draw a card.
//	 Whenever chaos ensues, you have no maximum hand size for the rest of the
//	 game."
//
// The first ability triggers once per creature that deals combat damage to a
// player and is controlled by the planar controller, but it is the creature's
// controller (read from the creature as it dealt the damage) who decides
// whether to draw (the rulings). The chaos ability is a player-level
// maximum-hand-size setting stamped at its resolution (CR 613.7b), so it lasts
// after the plane is gone and is ordered by timestamp against other
// hand-size effects, as the ruling asks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3f5a033-d063-4250-9cfe-4c3c1a2725f1",
		Name:         "Undercity Reaches",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, anyCreatureDealtCombatDamageToAPlayer,
				"Undercity Reaches — the creature's controller may draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					who := ctx.Controller()
					if o := ctx.Trigger().Object; o != nil && o.Controller != uuid.Nil {
						who = o.Controller
					}
					return MayChoice{
						Player:   who,
						Question: "Undercity Reaches — draw a card?",
						YesLabel: "Draw a card",
						NoLabel:  "Don't draw",
						Draws:    1,
						Then:     undercityReachesDraw,
					}.Apply(ctx)
				}),
			WheneverChaosEnsues("Undercity Reaches — you have no maximum hand size for the rest of the game",
				func(g *game.Game, item *game.StackItem) error {
					return g.SetMaxHandSizeForEffect(item.Controller, game.NoMaxHandSize)
				}),
		},
	})
}

// undercityReachesDraw is the creature's controller drawing on a yes. An
// on-disk key: never renamed, never reused.
var undercityReachesDraw = MayYes("confirm/undercity-reaches-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	return DrawCards{Player: r.Chooser, N: 1}.Apply(ctx)
})

// anyCreatureDealtCombatDamageToAPlayer is "a creature deals combat damage to
// a player", whoever controls either.
func anyCreatureDealtCombatDamageToAPlayer(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.IsCreature()
}
