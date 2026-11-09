package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flitterwing Nuisance — {U} Creature — Faerie Rogue 2/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Flying
//	 This creature enters with a -1/-1 counter on it.
//	 {2}{U}, Remove a counter from this creature: Whenever a creature you
//	 control deals combat damage to a player or planeswalker this turn,
//	 draw a card."
//
// The entry counter is Glen Elendra Guardian's replacement, and "a
// counter" is one counter of a kind the activator names (Loch Mare's
// RemoveCountersFromThis("", 1)), so a +1/+1 counter it picked up later
// pays too. The ability creates a CR 603.7b repeating delayed trigger
// (Great Train Heist's shape) that lasts the turn and is not tied to the
// Nuisance: it fires for each creature its controller controls that deals
// combat damage to a player or to a planeswalker, once per creature per
// combat damage step, even after the Nuisance has left. Activating it
// twice makes two such triggers, and each draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9148f2cb-5594-463c-a21b-a67706ec46d1",
		Name:            "Flitterwing Nuisance",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 1, "Flitterwing Nuisance: enters with a -1/-1 counter"),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{U}, Remove a counter from this creature: Whenever a creature you control deals combat damage to a player or planeswalker this turn, draw a card.",
			Cost:  Plus(ManaCost("{2}{U}"), RemoveCountersFromThis("", 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DelayedOnEvent{
					Label:     "Flitterwing Nuisance — draw a card",
					On:        []game.EventKind{game.EventDealDamage},
					Condition: flitterwingYourCreatureHitAPlayerOrPlaneswalker,
					Body:      drawOneBody,
					Repeats:   true,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// flitterwingYourCreatureHitAPlayerOrPlaneswalker is "whenever a creature
// you control deals combat damage to a player or planeswalker": a combat
// damage event from a creature the delayed trigger's controller controls
// as it dealt it (CR 603.7d), to a player or to a planeswalker. An on-disk
// key.
var flitterwingYourCreatureHitAPlayerOrPlaneswalker = game.DelayedCondition("combat/your-creature-hit-a-player-or-planeswalker",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
			return false
		}
		src, ok := g.LookupCardForEffect(ev.Source)
		if !ok || !src.IsCreature() || b17DamageSourceController(ev, g) != dt.Controller {
			return false
		}
		if g.PlayerByIDForEffect(ev.Target) != nil {
			return true
		}
		c, ok := g.LookupCardForEffect(ev.Target)
		return ok && c.IsPlaneswalker()
	})
