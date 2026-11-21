package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aberrant — Creature — Tyranid Mutant {X}{1}{G}, 0/0:
//
//	"Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or
//	 more, draw a card when it enters.)
//	 Trample
//	 Heavy Power Hammer — Whenever this creature deals combat damage to a
//	 player, destroy target artifact or enchantment that player controls."
//
// Ravenous is Mawloc's two halves (CR 702.156a): XCounters on entry and an
// intervening-if enters trigger that reads the X announced for the spell. The
// hammer's target is chosen among that player's artifacts and enchantments.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "46b5f12e-9fe5-44f6-b895-ee8c33205521",
		Name:                       "Aberrant",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"trample"},
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && source.CastX() >= 5
			}, "Aberrant — ravenous: X was 5 or more, draw a card", effects.Do(effects.DrawCards{N: 1})),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				Key:     "Aberrant — destroy target artifact or enchantment that player controls",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Source == source.InstanceID && effects.CombatDamageToPlayerBy(ev, source.Controller, g)
				},
				Targets: effects.ControlledByThatPlayer(effects.TargetPermanent("target artifact or enchantment that player controls", effects.Or(effects.Artifact(), effects.Enchantment()))),
				Effect:  effects.DestroyFirstLegalTarget,
			},
		},
	})
}
