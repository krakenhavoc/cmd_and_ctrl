package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_cost_variants.go — #6208: three activation-cost clauses
// the engine could not express, each a variant of a component it
// already had.
//
//	Illuminated Folio   "{1}, {T}, Reveal two cards from your hand
//	                     that share a color: Draw a card."
//	Tomb of Urami       "{2}{B}{B}, {T}, Sacrifice all lands you
//	                     control: Create Urami, …"
//	Benthic Explorers   "{T}, Untap a tapped land an opponent
//	                     controls: Add one mana of any type that land
//	                     could produce."
//
// The engine halves are game/reveal_cards_cost.go
// (RevealCardsCost.SharesColor), game/activated.go
// (AbilityCost.SacrificeAll, paid by game/sacrifice_all_cost.go) and
// game/tap_others_cost.go (TapOthersCost.OpponentControls). This file is
// the constructors and the Register guards.
//
// All three are DATA on a component that already existed, so the ADR
// 0041 closure ratchet gains no route from Game to a func.

// RevealNSharingAColor is "Reveal N cards from your hand that share a
// color" as the cost of an activated ability (#6208, CR 202.2) —
// Illuminated Folio's
//
//	Plus(ManaCost("{1}"), TapCost(), RevealNSharingAColor(2, "two cards from your hand that share a color"))
//
// The label is the clause as printed without the verb; the client's
// picker reads it. Revealing moves nothing (CR 701.20b), so the cards
// stay in hand.
//
// Every card named must have a colour in common with every other one,
// which for the printed count of two is simply "the two share a
// colour". A colourless card shares none (CR 105.2c), so it can never
// pay, and a hand that holds no sharing pair cannot activate the
// ability at all (CR 118.3). The clause names no quality of its own:
// the relation IS the clause, so Register refuses a colour or a
// colourless narrowing beside it.
func RevealNSharingAColor(n int, label string) game.AbilityCost {
	return game.AbilityCost{RevealCards: &game.RevealCardsCost{
		N:           n,
		Label:       label,
		SharesColor: true,
	}}
}

// SacrificeAllYouControl is "Sacrifice all <permanents> you control" as
// part of an activation cost (#6208) — Tomb of Urami's
//
//	Plus(ManaCost("{2}{B}{B}"), TapCost(), SacrificeAllYouControl("lands you control", Land()))
//
// SacrificeAllCost (#2097) one cost site over: the activator chooses
// nothing, and the engine takes every permanent they control that
// matches, as the ability is activated (CR 601.2b, 601.2h via CR
// 602.2b). None is a legal payment — an empty board pays the cost in
// full, since CR 118.3 asks only for what the cost needs. A phased-out
// permanent is not taken (CR 702.26b); an indestructible one is (CR
// 701.21a). The source goes with them when the clause matches it, as
// Tomb of Urami's own land does.
//
// The clause carries the "any number" bounds, so every reader that asks
// how many a payment may name already allows the whole board and zero;
// the flag says the number is "all of them".
func SacrificeAllYouControl(label string, preds ...CardPredicate) game.AbilityCost {
	return game.AbilityCost{
		SacrificeOther: SacrificeSpec(label, preds...).WithCount(0, 0),
		SacrificeAll:   true,
	}
}

// UntapTappedAnOpponentControls is "Untap N tapped <permanents> an
// opponent controls" as a COST (#6208, game.TapOthersCost.Untap with
// OpponentControls) — Benthic Explorers'
//
//	UntapTappedAnOpponentControls(1, "a tapped land an opponent controls", Land())
//
// UntapTapped with the owner turned over. Only a TAPPED permanent pays
// (CR 701.26b), one that can't become untapped does not (CR 614.17b),
// and a stun counter is removed in place of the untap with the cost
// still paid (CR 122.1d, CR 118.11). It does not target, so an
// opponent's hexproof land still pays (CR 601.2h, CR 602.2b), and it is
// not the {Q} symbol, so summoning sickness does not apply (CR 302.6).
//
// `label` is the clause as printed without the verb. The filter need not
// carry "an opponent controls": the engine's own owner check is what
// decides it, so a card file cannot widen or narrow it by forgetting a
// predicate.
//
// It may be the cost of a MANA ability as well as a CR 602 one, which
// the plain untap form may not: the activator never controls the
// permanent, so the source can never be a candidate and the
// "source pays both components" ordering (CR 601.2h) never arises.
// Which permanent was untapped is on PaidCost.UntappedOthers, which is
// how "one mana of any type that land could produce" is answered.
func UntapTappedAnOpponentControls(n int, label string, preds ...CardPredicate) *game.TapOthersCost {
	return &game.TapOthersCost{
		Count:            n,
		Filter:           TargetPermanent(label, preds...),
		Label:            label,
		Untap:            true,
		OpponentControls: true,
	}
}

// CheckActivationCostVariants is Register's guard for the three #6208
// clauses. Each refused shape compiles and then pays something no card
// prints.
func CheckActivationCostVariants(spec Spec) {
	checkRevealSharesColor(spec)
	checkAbilitySacrificeAll(spec)
	checkUntapFromOpponents(spec)
}

// checkRevealSharesColor holds the shared-colour reveal to the shape the
// engine reads: a fixed count of two or more on an activated ability,
// with no quality of its own. A count of one would be a clause every
// card pays (one card shares a colour with itself only when it has one),
// and the X form would ask an announcement for a number no card prints.
func checkRevealSharesColor(spec Spec) {
	for i, ab := range spec.Activated {
		rc := ab.Cost.RevealCards
		if rc == nil || !rc.SharesColor {
			continue
		}
		where := fmt.Sprintf("ability %d", i)
		switch {
		case rc.CountFromX:
			panic(fmt.Sprintf("effects.Register: %q %s reveals X cards that share a color — the clause has a printed count (RevealNSharingAColor, #6208)", spec.Name, where))
		case rc.N < 2:
			panic(fmt.Sprintf("effects.Register: %q %s reveals %d cards that share a color — a sharing clause names at least two (#6208)", spec.Name, where, rc.N))
		case rc.Color != "" || rc.Colorless:
			panic(fmt.Sprintf("effects.Register: %q %s reveals cards that share a color AND narrows their colour — the relation is the whole clause (#6208)", spec.Name, where))
		case rc.Creature || rc.Subtype != "" || rc.CardType != "":
			panic(fmt.Sprintf("effects.Register: %q %s reveals cards that share a color AND narrows their type — no printed card does (#6208)", spec.Name, where))
		}
	}
}

// checkAbilitySacrificeAll holds the ability's sacrifice-all clause to
// the shape sacrificeAllPaymentLocked fills: the open "any number"
// bounds and no choice of any kind in the clause. There is nothing to
// refuse on a MANA ability: ManaAbilityShape carries no such flag, so a
// clause put there stays the one-permanent pick it was.
func checkAbilitySacrificeAll(spec Spec) {
	for i, ab := range spec.Activated {
		if !ab.Cost.SacrificeAll {
			continue
		}
		where := fmt.Sprintf("ability %d", i)
		switch {
		case ab.Cost.SacrificeOther == nil:
			panic(fmt.Sprintf("effects.Register: %q %s sets SacrificeAll with no sacrifice clause — build it with SacrificeAllYouControl (#6208)", spec.Name, where))
		case !game.SacrificeAnyNumber(ab.Cost.SacrificeOther):
			panic(fmt.Sprintf("effects.Register: %q %s sets SacrificeAll on a clause with a count — build it with SacrificeAllYouControl, whose bounds are \"any number\" (#6208)", spec.Name, where))
		case len(ab.Cost.SacrificeOther.EachOf) > 0 || ab.Cost.SacrificeOther.Different != nil:
			panic(fmt.Sprintf("effects.Register: %q %s sets SacrificeAll on a clause with a set rule — \"all\" chooses nothing (#6208)", spec.Name, where))
		case ab.Cost.SacrificeSelf:
			panic(fmt.Sprintf("effects.Register: %q %s sacrifices all AND sacrifices its source — the clause already takes the source when it matches (#6208)", spec.Name, where))
		}
	}
}

// checkUntapFromOpponents keeps the opponent form (#6208) beside the
// untap form it is a reading of, and on an ability's cost: every other
// owner of a TapOthersCost pays through the tap payer, where the flag
// would compile and then tap a permanent its own validator refuses.
// checkUntapOthersCost (#5669) already refuses the plain untap form
// everywhere but a CR 602 cost; this allows the opponent form on a mana
// ability too, which ActivateManaAbility pays.
func checkUntapFromOpponents(spec Spec) {
	refuse := func(where string, tc *game.TapOthersCost) {
		if tc != nil && tc.OpponentControls {
			panic(fmt.Sprintf("effects.Register: %q %s untaps a permanent an opponent controls — only an ability's cost may (UntapTappedAnOpponentControls, #6208)", spec.Name, where))
		}
	}
	bad := func(where string, tc *game.TapOthersCost) {
		if tc == nil || !tc.OpponentControls {
			return
		}
		switch {
		case !tc.Untap:
			panic(fmt.Sprintf("effects.Register: %q %s TAPS a permanent an opponent controls — only the untap clause reaches another player's board (#6208)", spec.Name, where))
		case tc.ExcludeSource:
			panic(fmt.Sprintf("effects.Register: %q %s untaps \"another\" permanent an opponent controls — the source is never a candidate (#6208)", spec.Name, where))
		case tc.Filter == nil || tc.Filter.CountFromX || tc.Count < 1:
			panic(fmt.Sprintf("effects.Register: %q %s untaps no fixed number of an opponent's permanents — build it with UntapTappedAnOpponentControls (#6208)", spec.Name, where))
		}
	}
	for i, ab := range spec.Activated {
		bad(fmt.Sprintf("ability %d", i), ab.Cost.TapOthers)
	}
	for i, ma := range spec.ManaAbilities {
		bad(fmt.Sprintf("mana ability %d", i), ma.Cost.TapOthers)
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
