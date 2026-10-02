package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// state_trigger_b_helpers.go — shared bodies for the second batch of
// ADR 0107 §1's state-trigger cards (#1858): the tide-counter pair
// (Homarid, Tidal Influence), Force Bubble's end step and The Millennium
// Calendar.
//
// Append-only, per the shared-vocabulary rule.

// removeAllCountersFromThis is "remove all <kind> counters from this
// <permanent>" as an ability's whole effect (Homarid's and Tidal
// Influence's state trigger, Force Bubble's end step). A source that has
// left has none to remove, and one that came back is a new object
// (AddCounter, #1432).
func removeAllCountersFromThis(kind string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok || !onBattlefield(g, item.SourceCardID) || c.Counters[kind] <= 0 {
			return nil
		}
		return AddCounter{Target: item.SourceCardID, Kind: kind, N: -c.Counters[kind]}.Apply(NewContext(g, item))
	}
}

// whileExactlyTideCounters is a layer-7c static that applies `dp`/`dt` to
// every creature `who` accepts while the source has exactly `n` tide counters
// on it — Homarid's "As long as there is exactly one tide
// counter on this creature, it gets -1/-1", Tidal Influence's "As long
// as there are exactly three tide counters on this enchantment, all blue
// creatures get +2/+0". The layer engine recomputes on every counter
// change, so the count is read as it stands.
func whileExactlyTideCounters(n int, who func(target, source *game.Card) bool, dp, dt int) game.StaticAbility {
	return game.StaticAbility{
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return source.Counters["tide"] == n && target.IsCreature() && who(target, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += dp
			c.Toughness += dt
		},
	}
}

// thisCreatureOnly is the `who` of a self-only whileExactlyTideCounters.
func thisCreatureOnly(target, source *game.Card) bool { return target.InstanceID == source.InstanceID }

// blueCreatures is the `who` of "all blue creatures", anyone's.
func blueCreatures(target, _ *game.Card) bool { return target.HasColor("U") }
