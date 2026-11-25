package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inventor's Goggles — Artifact — Equipment {1} (Kaladesh, S82):
//
//	"Equipped creature gets +1/+2.
//	 Whenever an Artificer you control enters, you may attach this
//	 Equipment to it.
//	 Equip {2}"
//
// Hero's Blade's shape: the entering creature is read through the layered
// view (a changeling is an Artificer), the attach is optional and does not
// target, and it fails quietly when the Goggles or the creature has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee32ab1a-1120-4c34-a538-7bde9eff5a6e",
		Name:         "Inventor's Goggles",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 2)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Artificer"))
			},
			Key: "Inventor's Goggles — attach to the Artificer",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: item.Trigger.Event.CardID})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Inventor's Goggles — attach to it?"},
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
