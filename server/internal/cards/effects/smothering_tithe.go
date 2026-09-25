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
// No simplification remains. The Treasure's "{T}, Sacrifice this
// artifact: Add one mana of any color" went live in S21 sub-PR 1 —
// TreasureToken carries it on Card.ManaAbilities, which
// ManaAbilitiesForCard prefers over the catalog (a token has no
// oracle ID to look up). The note here claiming it was inert
// outlived the fix; corrected in the #338 sweep.
func init() {
	Register(Spec{
		OracleID:     "153376c9-dffd-458c-8ce3-a4c8269bc4e9",
		Name:         "Smothering Tithe",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller
			},
			Key: "Smothering Tithe — Treasure unless drawer pays {2}",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return PayUnless{
					Chooser:  ctx.Trigger().Event.Actor,
					Cost:     "{2}",
					Question: "Smothering Tithe — pay {2}?",
					OnDecline: func(ctx *Context) error {
						return CreateToken{
							Controller: ctx.Controller(),
							Template:   TreasureToken(),
							N:          1,
						}.Apply(ctx)
					},
				}.Apply(ctx)
			},
		}},
	})
}
