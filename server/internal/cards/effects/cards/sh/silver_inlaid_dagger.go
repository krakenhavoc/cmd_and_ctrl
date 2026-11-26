package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silver-Inlaid Dagger — {1} Artifact — Equipment (Innistrad, S82 OLD b448,
// #5572):
//
//	"Equipped creature gets +2/+0.
//	 As long as equipped creature is a Human, it gets an additional +1/+0.
//	 Equip {2}"
//
// Sharpened Pitchfork's shape: a plain +2/+0 modify and the same conditional
// +1/+0 read off the host every layer pass. "Human" is a word of the Dagger's
// text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93e9b0d6-97a8-4357-bcd4-f82d81bad232",
		Name:         "Silver-Inlaid Dagger",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			{
				Label:    "As long as equipped creature is a Human, it gets an additional +1/+0.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return effects.AttachedToSource(target, g, source) && target.HasSubtype(source.TextSubtype("Human"))
				},
				Apply: effects.PumpBy(1, 0),
			},
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
