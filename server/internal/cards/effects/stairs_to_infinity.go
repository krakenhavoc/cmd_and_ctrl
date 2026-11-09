package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stairs to Infinity — Plane — Xerex (Planechase Anthology, #4870):
//
//	"Players have no maximum hand size.
//	 Whenever you roll the planar die, draw a card.
//	 Whenever chaos ensues, reveal the top card of your planar deck. You
//	 may put it on the bottom of your planar deck."
//
// A plane (ADR 0216 §6). The hand size static is Anvil of Bogardan's. The
// draw watches the planar die's EventRollDie (any face, CR 901.3a) rolled
// by the planar controller. The chaos ability reveals the top of the
// communal planar deck (CR 901.15c) and asks whether it goes to the bottom;
// declined, it is turned face down on top again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a425df4-7010-462e-9cb5-20389bac721c",
		Name:         "Stairs to Infinity",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{PlayersHaveNoMaxHandSize()},
		Triggered: []game.TriggeredAbility{
			On(game.EventRollDie, youRolledThePlanarDie, "Stairs to Infinity — draw a card", Do(DrawCards{N: 1})),
			WheneverChaosEnsues("Stairs to Infinity — reveal the top card of your planar deck; you may put it on the bottom",
				func(g *game.Game, item *game.StackItem) error {
					top, ok := g.RevealTopOfPlanarDeckForEffect(item.Controller, item.SourceCardID, "Stairs to Infinity — the top card of the planar deck")
					if !ok {
						return nil
					}
					return MayChoice{
						Question: "Stairs to Infinity — put " + top.Name + " on the bottom of the planar deck?",
						Then:     stairsToInfinityBottom,
						Carry:    []uuid.UUID{top.InstanceID},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// youRolledThePlanarDie is "you roll the planar die": a planar die roll
// (any face) by the source's controller.
func youRolledThePlanarDie(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.PlanarFace != "" && ev.Actor == source.Controller
}

// stairsToInfinityBottom is the keyed answer. An on-disk key.
var stairsToInfinityBottom = MayYesNo("confirm/stairs-to-infinity-bottom",
	func(ctx *Context, r game.ConfirmAnswer) error {
		if len(r.Carry) > 0 {
			ctx.Game.PutTopOfPlanarDeckOnBottomForEffect(ctx.Controller(), r.Carry[0])
		}
		return nil
	},
	func(ctx *Context, r game.ConfirmAnswer) error {
		if len(r.Carry) > 0 {
			ctx.Game.HideTopOfPlanarDeckForEffect(ctx.Controller(), r.Carry[0])
		}
		return nil
	})
