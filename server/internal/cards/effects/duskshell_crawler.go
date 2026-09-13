package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duskshell Crawler — Creature — Insect {1}{G}, 0/3 (EDHREC rank
// 1417):
//
//	"When this creature enters, put a +1/+1 counter on target creature.
//	 Each creature you control with a +1/+1 counter on it has trample.
//	 (It can deal excess combat damage to the player or planeswalker
//	 it's attacking.)"
//
// The counters deck's two-drop: a counter on entry (any creature,
// itself included — a real target, picked when the trigger goes on
// the stack) and a Layer 6 trample grant over every creature its
// controller controls that carries a +1/+1 counter. The grant reads
// the counter map, which is instance state rather than a
// characteristic, so it is safe inside a layer recompute; the layer
// engine already recomputes on every counter change, so a creature
// gains trample the moment its first counter lands and loses it when
// the last comes off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e54ad5a1-e79d-42db-a3e9-5caecae62c9f",
		Name:         "Duskshell Crawler",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.Counters["+1/+1"] > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, k := range c.Abilities {
					if k == "trample" {
						return
					}
				}
				c.Abilities = append(c.Abilities, "trample")
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCreature("target creature"),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Duskshell Crawler — a +1/+1 counter on target creature",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return AddCounter{Target: item.Targets[0].ID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
