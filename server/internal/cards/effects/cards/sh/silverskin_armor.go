package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silverskin Armor — Artifact — Equipment {2} (Mirrodin Besieged, S82 OLD
// b466, #5576):
//
//	"Equipped creature gets +1/+1 and is an artifact in addition to its other
//	 types.
//	 Equip {2}"
//
// The pump is the layer 7c attached pump; the type is a layer 4 addition to
// the card types (it keeps every type it has), so the creature counts for
// metalcraft and artifact removal while it is equipped and stops when the
// Armor is unattached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4eb24ae3-da1a-4565-934f-e25f56036d71",
		Name:         "Silverskin Armor",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			{
				Label:     "Equipped creature is an artifact in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasType(c.Types, "Artifact") {
						c.Types = append(c.Types, "Artifact")
					}
				},
			},
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
