package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_built_test.go — #2884 follow-up (owner answer
// (a), 2026-10-09): a row with a Build fill-in is read by its declared
// Effect's footprint, and its item is accepted only when the Build set
// nothing the footprint does not read.

const (
	BuiltPlainProbe  = "test-2884-built-plain-probe"
	BuiltParamsProbe = "test-2884-built-params-probe"
)

func init() {
	plain := effects.WheneverYouCast(effects.Noncreature(), "Built Plain Probe — you gain 1 life", effects.Do(effects.GainLife{Amount: 1}))
	plain.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, "Built Plain Probe — you gain 1 life")
		item.Controller, item.Owner = ev.Actor, ev.Actor
		return item
	}
	effects.Register(effects.Spec{OracleID: BuiltPlainProbe, Name: "Built Plain Probe", Triggered: []game.TriggeredAbility{plain}})

	params := effects.WheneverYouCast(effects.Noncreature(), "Built Params Probe — you gain 1 life", effects.Do(effects.GainLife{Amount: 1}))
	params.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, "Built Params Probe — you gain 1 life")
		item.Params.Amount = 3
		return item
	}
	effects.Register(effects.Spec{OracleID: BuiltParamsProbe, Name: "Built Params Probe", Triggered: []game.TriggeredAbility{params}})
}
