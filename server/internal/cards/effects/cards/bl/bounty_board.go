package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bounty Board — Artifact {3}:
//
//	"{T}: Add one mana of any color.
//	 {1}, {T}: Put a bounty counter on target creature. Activate only as
//	 a sorcery.
//	 Whenever a creature with a bounty counter on it dies, each of its
//	 controller's opponents draws a card and gains 2 life."
//
// Mathas, Fiend Seeker's bounty counter, but the trigger is the Board's:
// it watches every creature that dies and asks the counters it had as it
// last existed (LastKnownCountersForEffect, CR 603.10a). "Its controller"
// is the player who controlled it as it died (leftUnderControlOf), and
// each of that player's opponents, the Board's controller among them,
// draws and then gains.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "e40c66cf-0a4a-4833-98bf-c685d35b9309",
		Name:          "Bounty Board",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.Cmdr2B01TapForAnyColor()},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, {T}: Put a bounty counter on target creature. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.AddCounter{Target: id, Kind: "bounty", N: 1}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if _, ok := effects.DiedCreature(ev, g); !ok {
					return false
				}
				return g.LastKnownCountersForEffect(ev.CardID)["bounty"] > 0
			}, "Bounty Board — each of its controller's opponents draws a card and gains 2 life",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					dead, ok := g.LookupCardForEffect(ev.CardID)
					if !ok {
						return nil
					}
					return effects.Cmdr2B01EachOpponentOfDrawsAndGains(g, item, effects.LeftUnderControlOf(ev, dead), 2)
				}),
		},
	})
}
