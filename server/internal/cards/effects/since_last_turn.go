package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// since_last_turn.go — the card-facing half of #5652: "if you haven't
// been dealt combat damage since your last turn" (Marchesa, Resolute
// Monarch). The engine half is game/since_last_turn.go, where
// Player.CombatDamagedSinceLastTurn is set as combat damage is dealt to
// a player and cleared as each of that player's turns ends.
//
//	AtYourUpkeepIfNoCombatDamageSinceYourLastTurn(label, Do(...))                 // the printed trigger
//	InterveningIf(CondNoCombatDamageSinceYourLastTurn{}, On(...))                 // the same "if" on another trigger

// CondNoCombatDamageSinceYourLastTurn is the intervening "if you haven't
// been dealt combat damage since your last turn", re-checked as the item
// resolves (CR 603.4). It declares
// game.FootprintCondNoCombatDamageSinceLastTurn: it can only become
// false, by combat damage, which no item on the stack deals (CR 510.2).
type CondNoCombatDamageSinceYourLastTurn struct{}

func (CondNoCombatDamageSinceYourLastTurn) Holds(ctx *Context) bool {
	return !ctx.Game.DealtCombatDamageSinceLastTurnForEffect(ctx.Controller())
}

func (CondNoCombatDamageSinceYourLastTurn) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondNoCombatDamageSinceLastTurn, 0), true
}

// NoCombatDamageSinceYourLastTurn is the same condition as the ability
// triggers: the source's controller has not been dealt combat damage
// since their last turn.
func NoCombatDamageSinceYourLastTurn(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return !g.DealtCombatDamageSinceLastTurnForEffect(source.Controller)
}

// AtYourUpkeepIfNoCombatDamageSinceYourLastTurn is "At the beginning of
// your upkeep, if you haven't been dealt combat damage since your last
// turn, <effect>": the condition is asked as the upkeep begins and again
// as the ability resolves (CR 603.4).
func AtYourUpkeepIfNoCombatDamageSinceYourLastTurn(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondNoCombatDamageSinceYourLastTurn{},
		On(game.EventBeginUpkeep, AllOf(ByYou, NoCombatDamageSinceYourLastTurn), label, effect))
}
