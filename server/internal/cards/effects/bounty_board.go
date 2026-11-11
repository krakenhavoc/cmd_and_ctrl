package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "e40c66cf-0a4a-4833-98bf-c685d35b9309",
		Name:          "Bounty Board",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{cmdr2B01TapForAnyColor()},
		Activated: []ActivatedAbility{{
			Label:        "{1}, {T}: Put a bounty counter on target creature. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}"), TapCost()),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return AddCounter{Target: id, Kind: "bounty", N: 1}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if _, ok := diedCreature(ev, g); !ok {
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
					return cmdr2B01EachOpponentOfDrawsAndGains(g, item, leftUnderControlOf(ev, dead), 2)
				}),
		},
	})
}
