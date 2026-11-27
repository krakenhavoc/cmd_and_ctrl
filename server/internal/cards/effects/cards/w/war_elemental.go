package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// War Elemental — Creature — Elemental {R}{R}{R}, 1/1 (Mirrodin, S82 OLD,
// #5606):
//
//	"When this creature enters, sacrifice it unless an opponent was dealt
//	 damage this turn.
//	 Whenever an opponent is dealt damage, put that many +1/+1 counters on
//	 this creature."
//
// The enters trigger always goes on the stack and the damage is read as it
// resolves (CR 608.2c, "unless" is not an intervening "if"): damage dealt
// to any opponent this turn, combat or not and from any source, saves it
// (AnOpponentWasDealtAtLeastThisTurn, Spinerock Knoll's reading). The
// second ability is one trigger per damage event to an opponent of the
// Elemental's controller, and puts counters equal to that event's amount,
// so 3 damage is three counters. Damage to its controller or to a
// permanent does nothing; damage that was prevented was never dealt.
//
// No simplification.
func init() {
	const name = "War Elemental"
	effects.Register(effects.Spec{
		OracleID:     "07947d01-ebb1-46b7-9bdd-ea00e034805b",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — sacrifice it unless an opponent was dealt damage this turn", func(g *game.Game, item *game.StackItem) error {
				if effects.AnOpponentWasDealtAtLeastThisTurn(1)(g, item.Controller) || !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			}),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target == uuid.Nil || ev.Target == source.Controller {
					return false
				}
				return g.PlayerByIDForEffect(ev.Target) != nil
			}, name+" — put that many +1/+1 counters on it", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return effects.CounterOnThis{Kind: "+1/+1", N: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
