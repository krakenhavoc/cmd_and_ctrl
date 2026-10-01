package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// state_trigger_b_helpers.go — shared bodies for the second batch of
// ADR 0107 §1's state-trigger cards (#1858): the tide-counter pair
// (Homarid, Tidal Influence), the damage-to-counters pair (Force Bubble,
// Nine Lives) and The Millennium Calendar.
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

// whileExactlyCounters is a layer-7c static that applies `dp`/`dt` to
// every creature `who` accepts while the source has exactly `n` counters
// of `kind` on it — Homarid's "As long as there is exactly one tide
// counter on this creature, it gets -1/-1", Tidal Influence's "As long
// as there are exactly three tide counters on this enchantment, all blue
// creatures get +2/+0". The layer engine recomputes on every counter
// change, so the count is read as it stands.
func whileExactlyCounters(kind string, n int, who func(target, source *game.Card) bool, dp, dt int) game.StaticAbility {
	return game.StaticAbility{
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return source.Counters[kind] == n && target.IsCreature() && who(target, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += dp
			c.Toughness += dt
		},
	}
}

// thisCreatureOnly is the `who` of a self-only whileExactlyCounters.
func thisCreatureOnly(target, source *game.Card) bool { return target.InstanceID == source.InstanceID }

// blueCreatures is the `who` of "all blue creatures", anyone's.
func blueCreatures(target, _ *game.Card) bool { return target.HasColor("U") }

// damageToYouBecomesCounters is the standing replacement of Force Bubble
// ("If damage would be dealt to you, put that many depletion counters on
// this enchantment instead") and Nine Lives ("If a source would deal
// damage to you, prevent that damage and put an incarnation counter on
// this enchantment"): damage to the source's controller is cancelled and
// counters go on the source — that many (`perDamage`), or one per damage
// event.
//
// The event is the damage to the PLAYER (CR 616.1: the affected player,
// the controller here, orders it against other replacements). Combat and
// noncombat alike. A source on which the counter cannot land still
// cancels the damage: the replacement applied, and CR 614.6 says the
// replaced event never happens.
func damageToYouBecomesCounters(kind string, perDamage bool, label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 &&
				ev.DamageTarget == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			n := 1
			if perDamage {
				n = ev.DamageAmount
			}
			ev.Cancel()
			return g.AddCounterForEffect(src.InstanceID, kind, n)
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}
