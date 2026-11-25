package cardsto

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// True-Faith Censer — Artifact — Equipment {2} (Shadows over Innistrad, #5543):
//
//	"Equipped creature gets +1/+1 and has vigilance.
//	 As long as equipped creature is a Human, it gets an additional +1/+0.
//	 Equip {2}"
//
// The extra +1/+0 is a layer-7c modify that reads the host's creature types
// as the earlier layers left them, so a creature that becomes a Human under
// the Censer (or stops being one) gains or loses it at once; a changeling
// is a Human.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "330f2e60-810d-4fe0-9a35-b3f2b29e3e2a",
		Name:         "True-Faith Censer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("vigilance"),
			{
				Label:     "As long as equipped creature is a Human, it gets an additional +1/+0.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if c.AllCreatureTypes || slices.Contains(c.Subtypes, "Human") {
						c.Power++
					}
				},
			},
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
