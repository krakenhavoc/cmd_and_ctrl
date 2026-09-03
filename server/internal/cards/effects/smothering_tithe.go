package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Smothering Tithe — Enchantment for {3}{W}:
//
//	"Whenever an opponent draws a card, that player may pay {2}. If
//	the player doesn't, you create a Treasure token."
//
// S19 sub-PR 6: an opponent-draws trigger. EventDrawCard fires once
// per card drawn with the drawing player in Actor, so "draw three"
// puts three Tithe triggers on the stack and asks three times —
// which is what the card does in paper. The trigger resolves into a
// PayUnless prompt for the drawer; declining hands the Tithe's
// controller a Treasure.
//
// Sandbox: the Treasure's sac-for-mana ability is inert until S21
// (see TreasureToken). Players who want to spend one today do it
// the S13 way — tap-and-track.
func init() {
	Register(Spec{
		OracleID: "153376c9-dffd-458c-8ce3-a4c8269bc4e9",
		Name:     "Smothering Tithe",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				drawer := ev.Actor
				return game.NewTriggeredItem(source, "Smothering Tithe — Treasure unless drawer pays {2}",
					func(g *game.Game, item *game.StackItem) error {
						return PayUnless{
							Chooser:  drawer,
							Cost:     "{2}",
							Question: "Smothering Tithe — pay {2}?",
							OnDecline: func(ctx *Context) error {
								return CreateToken{
									Controller: ctx.Controller(),
									Template:   TreasureToken(),
									N:          1,
								}.Apply(ctx)
							},
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
