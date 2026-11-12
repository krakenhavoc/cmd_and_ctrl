package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_tapped_cost.go — #5669: "Untap N tapped <permanents> you
// control" as a component of an activated ability's cost.

// UntapTapped is "Untap <n> tapped <permanents> you control" as a COST
// (#5669, game.TapOthersCost.Untap) — Halo Fountain's
//
//	UntapTapped(2, "two tapped creatures you control", Creature())
//
// The label is the clause as printed without the verb: the client's
// picker says "Untap <label> to pay for this ability". Compose it with
// the rest of the cost by setting the returned component on
// AbilityCost.TapOthers.
//
// It is the tap-others clause turned over: only a TAPPED permanent pays
// (CR 701.26b), one that can't become untapped does not (CR 614.17b),
// and a stun counter is removed in place of the untap with the cost
// still paid (CR 122.1d, CR 118.11). It is not the {Q} symbol, so
// summoning sickness does not apply (CR 302.6), and it does not target.
// When the same cost prints {T} and the source matches, the source may
// be named for both, in either order (CR 601.2h).
func UntapTapped(n int, label string, preds ...CardPredicate) *game.TapOthersCost {
	return &game.TapOthersCost{
		Count:  n,
		Filter: TargetPermanent(label, preds...),
		Label:  label,
		Untap:  true,
	}
}

// checkUntapOthersCost keeps the untap form (#5669) where the engine
// reads it: an activated ability's cost, with a fixed count. Every other
// owner of a TapOthersCost (a mana ability, a spell's additional,
// optional, alternative or splice cost) pays through the tap payer, so
// the flag there would compile and then tap nothing it could find.
func checkUntapOthersCost(spec Spec) {
	refuse := func(where string, tc *game.TapOthersCost) {
		if tc != nil && tc.Untap {
			panic(fmt.Sprintf("effects.Register: %q %s untaps tapped permanents — only an activated ability's cost may (#5669)", spec.Name, where))
		}
	}
	for i, ab := range spec.Activated {
		if tc := ab.Cost.TapOthers; tc != nil && tc.Untap && (tc.Filter == nil || tc.Filter.CountFromX || tc.Count < 1) {
			panic(fmt.Sprintf("effects.Register: %q ability %d untaps no fixed number of permanents — build it with UntapTapped (#5669)", spec.Name, i))
		}
	}
	for i, ma := range spec.ManaAbilities {
		refuse(fmt.Sprintf("mana ability %d", i), ma.Cost.TapOthers)
	}
	if ac := spec.AdditionalCost; ac != nil {
		refuse("additional cost", ac.TapOthers)
		for i, b := range ac.Either {
			refuse(fmt.Sprintf("either/or branch %d", i), b.TapOthers)
		}
	}
	for _, oc := range spec.OptionalCosts {
		refuse(fmt.Sprintf("optional cost %q", oc.Key), oc.TapOthers)
	}
	for _, alt := range spec.AlternativeCosts {
		refuse(fmt.Sprintf("alternative cost %q", alt.Key), alt.TapOthers)
	}
	if sp := spec.Splice; sp != nil && sp.Cost != nil {
		refuse("splice cost", sp.Cost.TapOthers)
	}
}
