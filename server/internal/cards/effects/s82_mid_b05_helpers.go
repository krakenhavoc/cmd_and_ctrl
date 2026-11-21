package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b05_helpers.go — Midnight Hunt coven cards, batch 5 (#5497).
// Coven itself is trekB48Coven (three or more creatures you control with
// different powers, read through the layers); these are the shapes the
// batch's cards share.

// midB05CovenCondition is the intervening "if you control three or more
// creatures with different powers" (CR 603.4), read as the trigger would
// go on the stack and again as it resolves.
func MidB05CovenCondition() Condition {
	return CondBoard{
		Name: "you control three or more creatures with different powers",
		Test: func(ctx *Context) bool {
			return ctx.Item != nil && TrekB48Coven(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID)
		},
	}
}

// midB05YouHaveCoven is the trigger-time half of the condition.
func MidB05YouHaveCoven(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return TrekB48Coven(g, source.Controller, source.InstanceID)
}

// midB05AtBeginningOfCombatIfCoven is "Coven — At the beginning of combat
// on your turn, if you control three or more creatures with different
// powers, <effect>".
func MidB05AtBeginningOfCombatIfCoven(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(MidB05CovenCondition(), On(game.EventStepBegan,
		AllOf(StepBegan(game.StepBeginCombat, true), MidB05YouHaveCoven), label, effect))
}

// midB05CovenThisGainsKeyword is the combat trigger "this creature gains
// <keyword> until end of turn" (Candlegrove Witch, Ritual Guardian,
// Candlelit Cavalry).
func MidB05CovenThisGainsKeyword(name, keyword string) game.TriggeredAbility {
	label := name + " — Coven: gains " + keyword + " until end of turn"
	return MidB05AtBeginningOfCombatIfCoven(label, ThisGainsKeywordUntilEndOfTurn(keyword, label))
}
