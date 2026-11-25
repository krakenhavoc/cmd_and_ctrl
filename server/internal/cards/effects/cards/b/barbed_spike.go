package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barbed Spike — Artifact — Equipment {1}{W} (Modern Horizons 2, #5501):
//
//	"When this Equipment enters, create a 1/1 colorless Thopter artifact
//	 creature token with flying, then attach this Equipment to it.
//	 Equipped creature gets +1/+0.
//	 Equip {2}"
//
// The create-then-attach rides one keyed tail (CreateTokenAndAttachSource),
// the Midnight Angel Armor shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28b22709-3e69-4eb7-8a3d-b75b6e65f30e",
		Name:         "Barbed Spike",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Barbed Spike — create a 1/1 colorless Thopter artifact creature token with flying, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenAndAttachSource(g, item, "1/1 colorless Thopter artifact with flying", true)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
