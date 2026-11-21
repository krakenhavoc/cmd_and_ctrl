package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Waterbender Ascension — Enchantment {1}{U} (Avatar: The Last
// Airbender, #4459):
//
//	"Whenever a creature you control deals combat damage to a player,
//	 put a quest counter on this enchantment. Then if it has four or more
//	 quest counters on it, draw a card.
//	 Waterbend {4}: Target creature can't be blocked this turn. (While
//	 paying a waterbend cost, you can tap your artifacts and creatures to
//	 help. Each one pays for {1}.)"
//
// The damage trigger is Bident of Thassa's: one per creature per player
// it connects with (combatDamageToPlayerBy). "Then if" is not an
// intervening "if" (CR 603.4): it is checked once, after the counter is
// put on, so the fourth hit already draws. An Ascension that has left
// gets no counter and draws nothing.
//
// WaterbendCost("{4}") is the whole cost (CR 701.67a); the target is
// re-checked as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	charged := effects.CondBoard{
		Name: "this enchantment has four or more quest counters on it",
		Test: func(ctx *effects.Context) bool { return waterbenderAscensionCharged(ctx.Game, ctx.Source()) },
	}
	effects.Register(effects.Spec{
		OracleID:     "aeddfdc2-149b-4703-ba57-8d844b25ef08",
		Name:         "Waterbender Ascension",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			}, "Waterbender Ascension — put a quest counter on it, then draw a card if it has four or more",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1}, effects.If{Cond: charged, Then: []effects.Applier{effects.DrawCards{N: 1}}})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Waterbend {4}: Target creature can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.WaterbendCost("{4}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.Msc17CantBeBlockedTheTarget("Waterbender Ascension — can't be blocked"),
		}},
	})
}

// waterbenderAscensionCharged is "it has four or more quest counters on
// it": false once it has left the battlefield.
func waterbenderAscensionCharged(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && effects.OnBattlefield(g, id) && c.Counters["quest"] >= 4
}
