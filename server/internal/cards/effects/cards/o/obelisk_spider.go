package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Obelisk Spider — Creature — Spider {1}{B}{G}, 1/4 (Hour of Devastation,
// #5534):
//
//	"Reach
//	 Whenever this creature deals combat damage to a creature, put a -1/-1
//	 counter on that creature.
//	 Whenever you put one or more -1/-1 counters on a creature, each opponent
//	 loses 1 life and you gain 1 life."
//
// The first trigger is Flayed Nim's combat-damage-to-a-creature shape: the
// damaged creature is recorded as the object it was (CR 400.7), and one that
// died to the damage or left gets nothing. The second watches counters
// placed by you (the placer is named on every placement, so a counter an
// opponent's effect put does not count): any -1/-1 counters put on any
// creature, one drain per placement, the Spider's own included, and the
// counter the first ability puts triggers it.
//
// No simplification.
func init() {
	const damagedLabel = "Obelisk Spider — put a -1/-1 counter on that creature"
	effects.Register(effects.Spec{
		OracleID:        "f7289e62-ce5e-4f11-8d64-9fb62531b578",
		Name:            "Obelisk Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventDealDamage},
				AppliesTo: effects.ThisDealtCombatDamageToACreature,
				Key:       damagedLabel,
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, damagedLabel)
					item.Params.Object = effects.DamagedCreatureRef(g, ev.Target)
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ref := item.Params.Object
					c, ok := g.LookupCardForEffect(ref.ID)
					if !ok || c.ObjectEpoch != ref.Epoch || !effects.OnBattlefield(g, ref.ID) {
						return nil
					}
					return effects.AddCounter{Target: ref.ID, Kind: game.CounterMinusOne, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCounterPlaced || ev.Label != game.CounterMinusOne ||
					effects.B11CounterPlacerOf(ev, g) != source.Controller || ev.Amount <= g.CounterCountBeforeForEffect(ev) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Target)
				return ok && c.IsCreature()
			}, "Obelisk Spider — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
