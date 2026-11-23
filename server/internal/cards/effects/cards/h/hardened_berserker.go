package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hardened Berserker — Creature — Human Berserker {2}{R}, 3/2:
//
//	"Whenever this creature attacks, the next spell you cast this turn
//	 costs {1} less to cast."
//
// The trigger goes on the stack and its resolution makes a one-use cost
// promise (#1852), priced in the CR 601.2f reduction pass and spent by
// the next spell cast, whatever it is. A second attack trigger this turn
// is a second promise, and two promises both apply to one spell.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2894efc6-ee78-4353-b98a-8005a2451139",
		Name:         "Hardened Berserker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source)
			}, "Hardened Berserker — the next spell you cast this turn costs {1} less", hardenedBerserkerPromise),
		},
	})
}

func hardenedBerserkerPromise(g *game.Game, item *game.StackItem) error {
	return effects.GrantNextSpellPromise{From: "Hardened Berserker", Promise: game.NextSpellPromise{
		Reduce: 1,
		Text:   "The next spell you cast this turn costs {1} less to cast.",
	}}.Apply(effects.NewContext(g, item))
}
