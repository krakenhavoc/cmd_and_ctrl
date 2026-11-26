package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB478LevelsAtLeast reports whether source has at least n level counters
// (CR 711.2b, "LEVEL N+"). A nil source has none.
func OldB478LevelsAtLeast(source *game.Card, n int) bool {
	return source != nil && source.Counters[game.CounterLevel] >= n
}

// OldB478CantBeBlockedExceptByAtLevel is a leveler band's "This creature can't
// be blocked except by <allowed>" (Zulaport Enforcer): a block rule on the
// creature itself that holds only while it has at least `level` level
// counters. A creature that has lost its abilities has lost the symbol and
// the rule with it, as the block-rule table reads it from the printed text.
func OldB478CantBeBlockedExceptByAtLevel(level int, allowed CardPredicate, label string) game.BlockRule {
	return CantBeBlockedExceptBy(func(g *game.Game, c, source *game.Card) bool {
		return OnSelf()(g, c, source) && OldB478LevelsAtLeast(source, level)
	}, allowed, label)
}

// OldB478PreventDamageToYouAndYourCreatures is a leveler band's "If a source
// would deal damage to you or a creature you control, prevent N of that
// damage" (Hedron-Field Purists, CR 615.1a), in force while the creature has
// from..to level counters (to == 0 is open-ended). Each band is one
// replacement; the bands never overlap, so one event is prevented once.
func OldB478PreventDamageToYouAndYourCreatures(name string, n, from, to int) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
				return false
			}
			lv := src.Counters[game.CounterLevel]
			if lv < from || (to != 0 && lv > to) {
				return false
			}
			if ev.DamageTarget == src.Controller {
				return true
			}
			t, ok := g.LookupCardForEffect(ev.DamageTarget)
			return ok && t.IsCreature() && t.Controller == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount -= n
			if ev.DamageAmount <= 0 {
				ev.Cancel()
			}
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      name + " — prevent " + roeB1Digits(n) + " of that damage",
	}
}
