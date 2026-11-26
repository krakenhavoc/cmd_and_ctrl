package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slumbering Dragon — Creature — Dragon {R}, 3/3 (Magic 2013):
//
//	"Flying
//	 This creature can't attack or block unless it has five or more +1/+1
//	 counters on it.
//	 Whenever a creature attacks you or a planeswalker you control, put a +1/+1
//	 counter on this creature."
//
// Sab-Sunen's conditional layer-6 CantAttack | CantBlock, read live off the
// +1/+1 counters on the Dragon itself, checked as attackers and blockers are
// declared (CR 508.1c, CR 509.1b); and Isperia's attack trigger, one per
// attacking creature (CR 508.3a), putting the counter on this creature only
// while it is still the object the trigger came from (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d424cab6-84be-4e55-8f4e-af11b1c50ede",
		Name:            "Slumbering Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label: "This creature can't attack or block unless it has five or more +1/+1 counters on it.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && source.Counters[game.CounterPlusOne] < 5
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil && effects.AttacksYouOrYourPlaneswalker(g, ev.Target, source.Controller)
			}, "Slumbering Dragon — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
