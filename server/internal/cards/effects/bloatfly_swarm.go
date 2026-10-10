package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloatfly Swarm — Creature — Insect Mutant {3}{B}, 0/0:
//
//	"Flying
//	 This creature enters with five +1/+1 counters on it.
//	 If damage would be dealt to this creature while it has a +1/+1
//	 counter on it, prevent that damage, remove that many +1/+1 counters
//	 from it, then give each player a rad counter for each +1/+1 counter
//	 removed this way."
//
// ADR 0108 §8 (#1906) is the prevention static and its additional
// effect (CR 615.5), Ugin's Conjurant's shape: "that many" is the damage,
// prevented or not (CR 615.12), or every counter it has when that is
// fewer. #2042 made the rad counters do something (CR 728.1), which is
// what held the card back. The rad counters follow the counters that
// actually came off, read before and after the removal, and the Swarm's
// controller gives them to every player still in the game, themselves
// included.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "fc50e254-71bc-46b3-93e6-c6eb820a3473",
		Name:            "Bloatfly Swarm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 5, "Bloatfly Swarm: enters with five +1/+1 counters"),
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				While: WhileItHasAPlusOneCounter,
				Then:  bloatflyRadBody,
				Label: "Bloatfly Swarm — prevent damage to it, remove that many +1/+1 counters and give each player that many rad counters",
			}),
		},
	})
}

// bloatflyRadBody is "remove that many +1/+1 counters from it, then give
// each player a rad counter for each +1/+1 counter removed this way". An
// on-disk key: never rename it.
var bloatflyRadBody = game.DelayedBody("prevention/bloatfly-remove-counters-give-rad", func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	info, ok := followUpThis(g, item)
	if !ok {
		return nil
	}
	before := info.Counters[game.CounterPlusOne]
	n := min(thatDamage(item), before)
	if n <= 0 {
		return nil
	}
	if err := g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, -n); err != nil {
		return err
	}
	removed := n
	if after, ok := followUpThis(g, item); ok {
		removed = before - after.Counters[game.CounterPlusOne]
	}
	return eachPlayerGetsRadCounters(g, item.Controller, removed)
})
