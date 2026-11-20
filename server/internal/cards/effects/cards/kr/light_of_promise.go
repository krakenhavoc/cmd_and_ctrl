package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Light of Promise — Enchantment — Aura {2}{W} (Marvel Universe, #4460):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever you gain life, put that many +1/+1
//	 counters on this creature.""
//
// Clever Distraction's shape: an Aura whose static grants the enchanted
// creature a triggered ability. The granted trigger belongs to the
// creature, so "you" is the creature's controller (a stolen creature pays
// its new controller) and "this creature" is the creature, not the Aura.
// It fires once per life-gain event (every gain path, lifelink included,
// emits a positive EventChangeLife) and puts as many counters as the
// life gained, read from the event as the trigger was queued. If the
// creature has left the battlefield when the trigger resolves, nothing
// happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64c54c9f-0550-41e2-a44a-5892a199f670",
		Name:         "Light of Promise",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants: []effects.AbilityGrant{{
			Key: lightOfPromiseGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverYouGainLife("Light of Promise — put that many +1/+1 counters on this creature",
					func(g *game.Game, item *game.StackItem) error {
						return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: item.Trigger.Event.Amount}.
							Apply(effects.NewContext(g, item))
					}),
			},
			Text: "Whenever you gain life, put that many +1/+1 counters on this creature.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(lightOfPromiseGrant)},
	})
}

// lightOfPromiseGrant is the key of the ability Light of Promise grants.
const lightOfPromiseGrant = "light-of-promise/gain-life-counters"
