package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ornery Dilophosaur — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "8291bc2e-ef58-4bd4-b261-b587142bbb1c", Name: "Ornery Dilophosaur", Completeness: CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{On(game.EventAttack, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			return ThisAttacked(ev, source, lki, g) && b24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
		}, "Ornery Dilophosaur — +2/+2 if you control a creature with power 4 or greater", Do(If{Cond: CondPowerAtLeast{N: 4}, Then: []Applier{Instruction(func(ctx *Context) error {
			return BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 2, Label: "Ornery Dilophosaur — +2/+2"}.Apply(ctx)
		})}}))},
	})
}
