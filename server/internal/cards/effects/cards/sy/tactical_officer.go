package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tactical Officer — Creature — Human Officer {1}{W}, 2/1 (Star Trek,
// #4973):
//
//	"Whenever you attack while creatures you control have total power 8
//	 or greater, put a +1/+1 counter on this creature. Untap it."
//
// The Chief Warg's shape: "whenever you attack" is once per declaration
// however many creatures attack (OncePerBatch), and the total power is a
// condition on the event, read as attackers are declared (Mosswort
// Bridge's sum, current power after the layers). The untap is of this
// creature, whether or not it attacked.
//
// No simplification.
func init() {
	atLeastEight := effects.CreaturesYouControlTotalPowerAtLeast(8)
	effects.Register(effects.Spec{
		OracleID:     "1397e7d5-4aa4-4c8d-8957-845579b775f7",
		Name:         "Tactical Officer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && atLeastEight(g, source.Controller)
			}, "Tactical Officer — put a +1/+1 counter on this creature, then untap it",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.PutCounterOnSelf(g, item); err != nil {
						return err
					}
					return effects.B31UntapSelf(g, item)
				})),
		},
	})
}
