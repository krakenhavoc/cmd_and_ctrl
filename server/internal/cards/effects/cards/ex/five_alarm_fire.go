package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Five-Alarm Fire — Enchantment {1}{R}{R} (Gatecrash, S82 OLD #5565):
//
//	"Whenever a creature you control deals combat damage, put a blaze counter
//	 on this enchantment.
//	 Remove five blaze counters from this enchantment: It deals 5 damage to any
//	 target."
//
// One trigger per creature that deals combat damage in a step, however many
// recipients it hurt (CR 510.2, CR 603.2c): the engine opens one damage event
// per recipient, so the trigger is OncePerBatch keyed by the dealing creature
// (a trampler hitting its blocker and the player is one trigger; two creatures
// are two; a double striker is one in each damage step, CR 510.4). The counter
// is skipped if the enchantment has left. The ability's cost removes five blaze
// counters (Obsidian Fireheart's kind), so it can't be activated with fewer;
// the enchantment is the damage source, as it last existed if it has left
// (CR 608.2h).
//
// No simplification.
func init() {
	const label = "Five-Alarm Fire — put a blaze counter on this enchantment"
	trigger := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
				return false
			}
			dealer, ok := g.LookupCardForEffect(ev.Source)
			return ok && dealer.IsCreature() && dealer.Controller == source.Controller
		},
		Key:          label,
		OncePerBatch: true,
		BatchKey:     func(ev game.Event, _ *game.Card, _ *game.Game) string { return ev.Source.String() },
		Effect: func(g *game.Game, item *game.StackItem) error {
			if !effects.OnBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
				return nil
			}
			return effects.AddCounter{Target: item.SourceCardID, Kind: "blaze", N: 1}.Apply(effects.NewContext(g, item))
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "b1e5c507-d110-4f9e-a58a-939f9961e464",
		Name:         "Five-Alarm Fire",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove five blaze counters from this enchantment: It deals 5 damage to any target.",
			Cost:    effects.RemoveCountersFromThis("blaze", 5),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 5)),
			Effect:  effects.DealDamageToTheTarget(5),
		}},
	})
}
