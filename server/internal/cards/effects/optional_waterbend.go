package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_waterbend.go — "As an additional cost to cast this spell, you
// may waterbend {N}." (CR 701.67, #3999), declared in Spec.OptionalCosts.
// The engine half is game/optional_waterbend.go: a chosen one charges its
// {N} like a kicker and becomes the cast's tap cost, so the caster's
// untapped artifacts and creatures may each pay {1} of it.
//
// A resolution asks whether it was paid with
// ctx.OptionalCostTimes(game.WaterbendKey) > 0.

// OptionalWaterbend is "you may waterbend {cost}" as an optional
// additional cost. `cost` is the waterbend's mana in brace notation,
// generic only ("{10}"), as every printed waterbend cost is.
func OptionalWaterbend(cost string) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:  true,
		Key:       game.WaterbendKey,
		Label:     "Waterbend " + cost,
		ManaCost:  cost,
		Waterbend: Waterbend(cost),
	}
}

// checkOptionalWaterbend refuses the shapes of a waterbend additional
// cost that would compile and then charge or tap something the card does
// not print:
//
//   - a Waterbend clause in the mandatory slot or on an either/or branch
//     (a mandatory waterbend is Spec.TapCost, Waterbender's Restoration);
//   - an optional one built by hand: another key, a mana cost that is not
//     the clause's own, or any other component beside it;
//   - one on a card that also has a tap cost of its own, or two of them:
//     a cast carries one tap cost and one tap_ids list;
//   - an optional cost keyed WaterbendKey with no clause, which would
//     charge the {N} and let nothing be tapped for it.
func checkOptionalWaterbend(spec Spec) {
	if ac := spec.AdditionalCost; ac != nil {
		if !ac.Waterbend.Empty() {
			panic(fmt.Sprintf("effects.Register: %q puts an optional waterbend in the mandatory AdditionalCost slot — a mandatory waterbend is Spec.TapCost = Waterbend(...); an optional one is OptionalWaterbend in OptionalCosts", spec.Name))
		}
		for _, br := range ac.Either {
			if !br.Waterbend.Empty() {
				panic(fmt.Sprintf("effects.Register: %q puts a waterbend on an either/or branch %q — not modelled (#3999)", spec.Name, br.Key))
			}
		}
	}
	seen := 0
	for _, oc := range spec.OptionalCosts {
		if oc.Waterbend.Empty() {
			if oc.Key == game.WaterbendKey {
				panic(fmt.Sprintf("effects.Register: %q declares an optional cost keyed %q with no waterbend clause — build it with OptionalWaterbend", spec.Name, oc.Key))
			}
			continue
		}
		seen++
		rest := oc
		rest.Waterbend, rest.ManaCost, rest.Key, rest.Label, rest.Optional = nil, "", "", "", false
		if oc.Key != game.WaterbendKey || oc.ManaCost == "" || oc.ManaCost != oc.Waterbend.Extra || oc.Waterbend.Key != "waterbend" ||
			oc.Repeat > 1 || oc.MinX != 0 || oc.Targets != nil || !rest.Empty() {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q is a malformed waterbend cost — build it with OptionalWaterbend", spec.Name, oc.Key))
		}
		if p, err := game.ParseCost(oc.ManaCost); err != nil || p.XSlots > 0 || len(p.Required) > 0 {
			panic(fmt.Sprintf("effects.Register: %q optional waterbend %q is not a fixed generic cost", spec.Name, oc.ManaCost))
		}
	}
	if seen > 1 {
		panic(fmt.Sprintf("effects.Register: %q declares %d optional waterbend costs — a cast carries one tap cost", spec.Name, seen))
	}
	if seen == 1 && spec.TapCost != nil {
		panic(fmt.Sprintf("effects.Register: %q has a tap cost of its own and an optional waterbend — a cast carries one tap cost", spec.Name))
	}
}
