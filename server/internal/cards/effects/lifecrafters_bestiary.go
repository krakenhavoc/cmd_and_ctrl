package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lifecrafter's Bestiary — Artifact {3} (EDHREC rank 1404):
//
//	"At the beginning of your upkeep, scry 1.
//	 Whenever you cast a creature spell, you may pay {G}. If you do,
//	 draw a card."
//
// The green creature deck's card-draw rock. The upkeep scry is the
// Scry primitive (a prompt, nothing moves until answered); the cast
// trigger is the MayPay primitive — a prompt for the controller that
// draws on "Pay" (with {G} in pool or from an untapped source) and
// does nothing on "Don't pay". Cast, not resolve, so a countered
// creature still offers the draw, and the spell's type is read off
// the stack where its type line is intact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f2a3797-28aa-4c7a-ba2b-fd243a1747fd",
		Name:         "Lifecrafter's Bestiary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginUpkeep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Lifecrafter's Bestiary — scry 1",
						func(g *game.Game, item *game.StackItem) error {
							return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventCast},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b12CreatureSpellCastByYou(ev, source, g)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Lifecrafter's Bestiary — pay {G} to draw a card",
						func(g *game.Game, item *game.StackItem) error {
							return MayPay{
								Chooser:  item.Controller,
								Cost:     "{G}",
								Question: "Lifecrafter's Bestiary — pay {G} to draw a card?",
								OnPay: func(ctx *Context) error {
									return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
								},
							}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
