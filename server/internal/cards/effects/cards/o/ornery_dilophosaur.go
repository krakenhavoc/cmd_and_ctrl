package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ornery Dilophosaur — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "8291bc2e-ef58-4bd4-b261-b587142bbb1c", Name: "Ornery Dilophosaur", Completeness: effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{effects.On(game.EventAttack, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			return effects.ThisAttacked(ev, source, lki, g) && effects.B24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
		}, "Ornery Dilophosaur — +2/+2 if you control a creature with power 4 or greater", effects.Do(effects.If{Cond: effects.CondPowerAtLeast{N: 4}, Then: []effects.Applier{effects.Instruction(func(ctx *effects.Context) error {
			return effects.BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 2, Label: "Ornery Dilophosaur — +2/+2"}.Apply(ctx)
		})}}))},
	})
}
