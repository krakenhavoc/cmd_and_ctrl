package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b11_helpers.go — shared bodies for Dominaria United batch 11.

// dmuB11TargetCard is the one card the trigger or spell targets, if it is
// still a legal target as the ability resolves (CR 608.2b).
func dmuB11TargetCard(ctx *Context) (uuid.UUID, bool) {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return t.ID, true
		}
	}
	return uuid.Nil, false
}

// dmuB11DomainCounters is "put X +1/+1 counters on target creature you
// control, where X is the number of basic land types among lands you
// control" (Briar Hydra). X is counted as the ability resolves.
func dmuB11DomainCounters() Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := dmuB11TargetCard(ctx)
		if !ok {
			return nil
		}
		n := DomainCount(g, item.Controller)
		if n == 0 {
			return nil
		}
		return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
	}
}

// dmuB11BecameTapped is the "becomes tapped" trigger shape: the tap
// event or an attack that taps the creature (ThisBecameTapped, #4534).
func dmuB11BecameTapped(label string, effect Effect) game.TriggeredAbility {
	return OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped, label, effect)
}

// dmuB11CountersOnSmallCreatures is Zar Ojanen's body: a +1/+1 counter on
// each creature the controller controls whose toughness is less than the
// domain count. The set is fixed before any counter is placed (CR 608.2),
// so a counter that lifts one creature never pulls another in.
func dmuB11CountersOnSmallCreatures(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := DomainCount(g, item.Controller)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.Toughness < n {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// DivideSourcePlusOneCounters is "distribute X +1/+1 counters among …,
// where X is the number of +1/+1 counters on this creature" on a dies
// trigger (Quirion Beastcaller): the counters it had as it last existed
// on the battlefield (CR 603.10, 608.2h), read as the trigger is put on
// the stack. A source with no last-known record reads its live counters.
var DivideSourcePlusOneCounters = game.RegisterDivideAmount("source-plus-one-counters",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		if info, ok := g.LastKnownPermanentForEffect(a.Source); ok {
			return game.DivideSpec{Total: info.Counters[game.CounterPlusOne]}
		}
		if c, ok := g.LookupCardForEffect(a.Source); ok {
			return game.DivideSpec{Total: c.Counters[game.CounterPlusOne]}
		}
		return game.DivideSpec{}
	})

// searchRuleDifferentPowers is "creature cards with different powers"
// (Threats Undetected): no two picked cards share a power. The power is the
// one the card has in the library, its printed value.
var searchRuleDifferentPowers = game.RegisterSearchRule("search/different-powers",
	func(picked []game.Card, _ int) bool {
		seen := make(map[int]bool, len(picked))
		for _, c := range picked {
			p := c.PowerForComparison()
			if seen[p] {
				return false
			}
			seen[p] = true
		}
		return true
	})
