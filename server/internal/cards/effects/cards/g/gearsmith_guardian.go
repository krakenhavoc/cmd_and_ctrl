package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gearsmith Guardian — Artifact Creature — Construct {5}, 3/5
// (Core Set 2019, #5527):
//
//	"This creature gets +2/+0 as long as you control a blue creature."
//
// A layer-7c pump on itself while you control a blue creature
// (Wingrattle Scarecrow's condition, selfWhileYouControlA).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c0f7092-3fe6-4a6e-8b4e-803810b43e50",
		Name:         "Gearsmith Guardian",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Gets +2/+0 as long as you control a blue creature.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.And(effects.Creature(), effects.OfColor("U"))),
			Apply:     effects.PumpBy(2, 0),
		}},
	})
}
