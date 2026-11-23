package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anti-Venom, Horrifying Healer — Legendary Creature — Symbiote Hero {W}{W}{W}{W}{W}, 5/5:
//
//	"When Anti-Venom enters, if he was cast, return target creature card from your graveyard to the battlefield.
//	 If damage would be dealt to Anti-Venom, prevent that damage and put that many +1/+1 counters on him."
//
// ADR 0108 §8 (#1906): a prevention static whose additional effect puts
// "that many" +1/+1 counters on him — the damage, prevented or not, so
// damage that can't be prevented is dealt and still adds the counters
// (CR 615.12). One application per recipient in a damage instance, with
// the instance's total.
//
// "If he was cast" is a CR 603.4 intervening if read off the permanent's
// own cast record (CR 400.7d), as Wedding Ring's is: Anti-Venom put onto
// the battlefield any other way returns nothing.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c7bafe9-80cd-48d0-bcae-e7910c9fb83b",
		Name:         "Anti-Venom, Horrifying Healer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.WeddingRingWasCast),
					"Anti-Venom — return target creature card from your graveyard to the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						effects.ReanimateSingleTarget(effects.NewContext(g, item), item.Controller)
						return nil
					}),
				effects.TargetCreatureInYourGraveyard()),
		},
		Replacements: []game.ReplacementEffect{
			effects.PreventDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToThisCreature,
				Then:  effects.ThatManyCountersOnThisBody,
				Label: "Anti-Venom — prevent damage to him and put that many +1/+1 counters on him",
			}),
		},
	})
}
