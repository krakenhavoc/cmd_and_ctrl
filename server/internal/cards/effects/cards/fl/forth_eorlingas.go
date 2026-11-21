package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forth Eorlingas! — Sorcery {X}{R}{W} (Tales of Middle-earth Commander):
//
//	"Create X 2/2 red Human Knight creature tokens with trample and
//	 haste. Whenever one or more creatures you control deal combat
//	 damage to one or more players this turn, you become the monarch."
//
// The second sentence is CR 603.7b's repeating delayed trigger, Great
// Train Heist's shape: the spell sets it up, it is not tied to the
// sorcery and ends in the cleanup step. It watches combat damage dealt
// to any player by a creature the caster controls (the condition is the
// shared combatDamageToPlayerBy read, so a creature that has left is
// judged by last-known information); it fires per damage event, and a
// second firing while the caster already wears the crown changes
// nothing (CR 725.3). The tokens are made first, so they can connect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35807c14-4647-482f-9ba1-6d640fd5e8d8",
		Name:         "Forth Eorlingas!",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{
				Controller: item.Controller,
				Template:   effects.TokenCard("2/2 red Human Knight with trample and haste"),
				N:          item.XValue,
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.DelayedOnEvent{
				Label:     "Forth Eorlingas! — you become the monarch",
				On:        []game.EventKind{game.EventDealDamage},
				Condition: forthEorlingasCondition,
				Body:      forthEorlingasBody,
				Repeats:   true,
			}.Apply(ctx)
		},
	})
}

var (
	forthEorlingasCondition = game.DelayedCondition("combat/your-creature-hit-a-player", func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		return effects.CombatDamageToPlayerBy(ev, dt.Controller, g)
	})
	forthEorlingasBody = game.SimpleDelayedBody("monarch/become", func(g *game.Game, item *game.StackItem) error {
		return effects.BecomeTheMonarch{}.Apply(effects.NewContext(g, item))
	})
)
