package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Duskana, the Rage Mother — Legendary Creature — Bear {2}{R}{G}{W}, 5/5
// (Murders at Karlov Manor Commander):
//
//	"When Duskana enters, draw a card for each creature you control with base
//	 power and toughness 2/2.
//	 Whenever a creature you control with base power and toughness 2/2
//	 attacks, it gets +3/+3 until end of turn."
//
// "Base power and toughness 2/2" is read after layer 7b (CR 208.4b), so a
// counter or an anthem does not stop a 2/2 counting and one does not make a
// 1/1 count. The draw counts as it resolves, Duskana's own 5/5 never counting.
// The attack trigger asks about the attacker as it is declared (CR 603.2);
// each such attacker triggers separately and the pump goes to that creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68d97ebb-63f8-4795-bddb-03582e13b34d",
		Name:         "Duskana, the Rage Mother",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Duskana, the Rage Mother — draw a card for each creature you control with base power and toughness 2/2",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := cmdr2B21BasePTCount(g, ctx.Controller(), cmdr2B21IsTwoByTwo)
					return DrawCards{N: n}.Apply(ctx)
				}),
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && cmdr2B21IsTwoByTwo(c.BasePower(), c.BaseToughness())
			}, "Duskana, the Rage Mother — the attacker gets +3/+3 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					id := item.Trigger.Event.CardID
					return BoostUntilEOT{Target: id, Power: 3, Toughness: 3,
						Label: "Duskana, the Rage Mother — +3/+3 until end of turn"}.Apply(NewContext(g, item))
				}),
		},
	})
}
