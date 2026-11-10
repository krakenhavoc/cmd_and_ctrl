package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b05_helpers.go — shared bodies for the Merfolk batch (#5368).

// cmdr2B05MerfolkFilter is "Merfolk creatures", every Merfolk or only
// the source controller's, and with or without the source itself.
func cmdr2B05MerfolkFilter(others, yours bool) TribeFilter {
	return TribeFilter{Tribes: []string{"Merfolk"}, Others: others, YoursOnly: yours}
}

// cmdr2B05ControlsOtherMerfolk reports whether the source's controller
// controls at least n Merfolk creatures other than the source. It reads
// the layered view on every recompute (CR 611.3a), changelings included.
func cmdr2B05ControlsOtherMerfolk(g *game.Game, source *game.Card, n int) bool {
	f := cmdr2B05MerfolkFilter(true, true)
	count := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		c := c
		if f.Matches(&c, g, source) {
			count++
		}
	}
	return count >= n
}

// cmdr2B05SelfKeywordWhileOtherMerfolk is "<this> has <keyword> as long
// as you control at least n other Merfolk".
func cmdr2B05SelfKeywordWhileOtherMerfolk(n int, keyword string) game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfCreature(target, g, source) && cmdr2B05ControlsOtherMerfolk(g, source, n)
	}, keyword)
	ab.Label = "Has " + keyword + " as long as you control at least " + numberWord(n) + " other Merfolk."
	return ab
}

// cmdr2B05TargetsYourMerfolk passes when one of the announced targets is
// a Merfolk creature the modifier source's controller controls — Kopala's
// "that target a Merfolk you control". A target-reading predicate: the
// modifier that uses it sets ReadsTargets.
func cmdr2B05TargetsYourMerfolk() CostPredicate {
	f := cmdr2B05MerfolkFilter(false, false)
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		for _, t := range q.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			c, ok := q.Game.LookupCardForEffect(t.ID)
			if ok && c.Controller == q.Source.Controller && f.Matches(&c, q.Game, &q.Source) {
				return true
			}
		}
		return false
	}
}

// cmdr2B05SpellOfAnyType passes on a spell of any of the named creature
// types — "Merfolk spells and Wizard spells". Read off the spell as it
// sits on the stack, where HasSubtype answers true for every creature
// type on a changeling (CR 702.73a); the names are words of the source's
// text (CR 612.2).
func cmdr2B05SpellOfAnyType(subtypes ...string) CostPredicate {
	return func(q game.CostQuery) bool {
		for _, s := range subtypes {
			if q.Card.HasSubtype(q.Source.TextSubtype(s)) {
				return true
			}
		}
		return false
	}
}

// cmdr2B05TapMerfolk is the "tap N untapped Merfolk you control" cost
// component (CR 602.2, game/tap_others_cost.go). `another` is the printed
// word "another": the source may not pay itself. Merfolk is an effective
// subtype, so a changeling pays.
func cmdr2B05TapMerfolk(n int, another bool, label string) game.AbilityCost {
	return game.AbilityCost{TapOthers: &game.TapOthersCost{
		Count:         n,
		ExcludeSource: another,
		Filter:        TargetCreature(label, HasSubtype("Merfolk")),
		Label:         label,
	}}
}

// cmdr2B05CountersOnEachMerfolk puts n +1/+1 counters on each Merfolk
// creature the item's controller controls, one placement per creature.
func cmdr2B05CountersOnEachMerfolk(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		f := cmdr2B05MerfolkFilter(false, true)
		src := game.Card{Controller: item.Controller}
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			c := c
			if f.Matches(&c, g, &src) {
				ids = append(ids, c.InstanceID)
			}
		}
		for _, id := range ids {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
				continue
			}
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}).Apply(ctx.asGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}
}
