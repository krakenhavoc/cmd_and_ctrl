package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3c7bafe9-80cd-48d0-bcae-e7910c9fb83b",
		Name:         "Anti-Venom, Horrifying Healer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, weddingRingWasCast),
					"Anti-Venom — return target creature card from your graveyard to the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						reanimateSingleTarget(NewContext(g, item), item.Controller)
						return nil
					}),
				targetCreatureInYourGraveyard()),
		},
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				Then:  thatManyCountersOnThisBody,
				Label: "Anti-Venom — prevent damage to him and put that many +1/+1 counters on him",
			}),
		},
	})
}
