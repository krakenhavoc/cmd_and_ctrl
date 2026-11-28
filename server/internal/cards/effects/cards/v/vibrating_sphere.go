package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vibrating Sphere — Artifact {4} (Ice Age, S82 OLD, #5636):
//
//	"During your turn, creatures you control get +2/+0.
//	 During turns other than yours, creatures you control get -0/-2."
//
// Two layer 7c statics read live (CR 611.3a), Street Riot's gate and its
// opposite, both over the creatures its controller controls.
//
// No simplification.
func init() {
	yours := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature()
	}
	onYourTurn := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return yours(target, g, source) && effects.IsActivePlayer(g, source.Controller)
	}
	onOtherTurns := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return yours(target, g, source) && !effects.IsActivePlayer(g, source.Controller)
	}
	plus := effects.B16Anthem(onYourTurn, 2, 0)
	plus.Label = "During your turn, creatures you control get +2/+0."
	minus := effects.B16Anthem(onOtherTurns, 0, -2)
	minus.Label = "During turns other than yours, creatures you control get -0/-2."
	effects.Register(effects.Spec{
		OracleID:     "8a8fe94f-81aa-4725-ace6-ed56a1e1ed91",
		Name:         "Vibrating Sphere",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{plus, minus},
	})
}
