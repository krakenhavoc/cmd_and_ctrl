package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kederekt Parasite — Creature — Horror {B}, 1/1 (EDHREC rank 1998):
//
//	"Whenever an opponent draws a card, if you control a red
//	 permanent, you may have this creature deal 1 damage to that
//	 player."
//
// The Nekusar deck's one-drop. An intervening-if trigger (CR
// 603.4): "if you control a red permanent" is checked when the
// opponent draws — no red permanent, no trigger, no prompt — and
// checked again as the trigger resolves, so a red permanent that
// left in response makes the ability do nothing. Once per card
// drawn, as printed: a wheel is one prompt per card. "That player"
// is the draw event's Actor, read at resolution off the item's
// carried trigger context (item.Trigger.Event.Actor, #1223); the
// damage comes from the Parasite, so it is Parasite damage whether or
// not it is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc7b46af-6c07-455a-99c7-f1bccaa2a5f4",
		Name:         "Kederekt Parasite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == uuid.Nil || ev.Actor == source.Controller {
					return false
				}
				return b18ControlsPermanentOfColor(g, source.Controller, "R")
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Kederekt Parasite — deal 1 damage to the player who drew?"},
			Key:            "Kederekt Parasite — 1 damage to the player who drew",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !b18ControlsPermanentOfColor(g, item.Controller, "R") {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
