package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ultros, Obnoxious Octopus — Legendary Creature — Octopus {1}{U}, 2/1
// (Final Fantasy, #4462):
//
//	"Whenever you cast a noncreature spell, if at least four mana was
//	 spent to cast it, tap target creature an opponent controls and put
//	 a stun counter on it.
//	 Whenever you cast a noncreature spell, if at least eight mana was
//	 spent to cast it, put eight +1/+1 counters on Ultros."
//
// Two separate cast triggers over Blazing Bomb's payment read, so an
// eight-mana spell sets off both. The first targets as it goes on the
// stack (CR 603.3d) and is removed with no creature an opponent controls
// to target; its body is Cryotheory Adept's tap-and-stun. The second
// puts its counters only on the Ultros it came from (CR 400.7).
//
// No simplification.
func init() {
	atLeast := func(n int) func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			return effects.YouCast(effects.Noncreature())(ev, source, lki, g) && g.StackItemPaidForEffect(ev.CardID).Spent().Total() >= n
		}
	}
	effects.Register(effects.Spec{
		OracleID:       "b6f4b524-1081-4263-b7e4-9aee93131b31",
		Name:           "Ultros, Obnoxious Octopus",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, atLeast(4),
					"Ultros — tap target creature an opponent controls and put a stun counter on it",
					effects.TapTargetAndStunIt),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
			effects.On(game.EventCast, atLeast(8), "Ultros — put eight +1/+1 counters on Ultros",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 8})),
		},
	})
}
