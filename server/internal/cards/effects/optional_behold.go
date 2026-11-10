package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_behold.go — "As an additional cost to cast this spell, you
// may behold a <type>." (CR 701.4a behold, #4642; Hulk's Thunderclap,
// Hyperion's Atomic Vision), declared in Spec.OptionalCosts.
//
// The component is BeholdCost's (game.RevealCost with Behold set): the
// caster who takes the offer names ONE card on cast_spell's reveal_ids,
// a <type> they control or a <type> card in their hand, and a card from
// hand is revealed to the table with the spell on the stack. The offer
// is an ordinary optional cost, so the announcement, the paid record
// (PaidCost.OptionalCosts, which snapshots and undo already carry), the
// view and the bot's move list are kicker's. "If a Hero was beheld" is a
// fact about the announcement (CR 701.4b), not about the card chosen, so
// the resolution asks ctx.WasBeheld() and never looks at the card.

// OptionalBehold is "As an additional cost to cast this spell, you may
// behold <article> <subtype>" as an optional additional cost. `article`
// is the printed "a" or "an" and `subtype` the creature type ("Hero",
// "Gamma"). A text that reads "behold a Gamma creature" passes "Gamma":
// the choice is a Gamma you control or a Gamma card in hand.
func OptionalBehold(article, subtype string) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      game.BeholdKey,
		Reveal:   &game.RevealCost{Subtype: subtype, Behold: true},
		Label:    "Behold " + article + " " + subtype,
	}
}

// OptionalBeholdChosenType is "As an additional cost to cast this spell,
// you may choose a creature type and behold <n> creatures of that type"
// (#5406, Celestial Reunion). The caster who takes the offer names the
// type on cast_spell's `behold_type` as they announce it (CR 601.2b) and
// exactly n objects of that type on `reveal_ids`: creature cards in hand
// and creatures they control, in any mix (CR 701.4a). Creatures only: a
// Kindred card or permanent of the type does not pay it. The type is on
// the stack item; the resolution reads it through ctx.BeheldType.
func OptionalBeholdChosenType(n int) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      game.BeholdKey,
		Reveal:   &game.RevealCost{Behold: true, ChosenType: true, Creature: true, Count: n},
		Label:    fmt.Sprintf("Choose a creature type and behold %s creatures of that type", numberWord(n)),
	}
}

// BeheldType is "the chosen type" of a behold that chose its creature
// type (OptionalBeholdChosenType, #5406): the type the caster named as
// they cast the spell (CR 601.2b), read off the stack item. ok is false
// when the offer was declined and for every ability item.
func (c *Context) BeheldType() (string, bool) {
	if c.Item == nil || c.Item.Paid.BeholdType == "" || !c.WasBeheld() {
		return "", false
	}
	return c.Item.Paid.BeholdType, true
}

// OptionalBeholdOpeningFlash is OptionalBehold whose announcement also
// lets the spell be cast as though it had flash: "You may cast this spell
// as though it had flash if you behold a <subtype> as an additional cost
// to cast it" (Molten Exhale, CR 601.3c, #5207).
func OptionalBeholdOpeningFlash(article, subtype string) game.AdditionalCost {
	c := OptionalBehold(article, subtype)
	c.AsThoughFlash = true
	return c
}

// WasBeheld is "if this spell's additional cost was paid" / "if a Hero
// was beheld" for a spell with an OptionalBehold cost: the caster took
// the offer when they cast it (CR 601.2b, 701.4b). False when they
// declined and for every ability item.
func (c *Context) WasBeheld() bool {
	return c.OptionalCostTimes(game.BeholdKey) > 0
}

// checkOptionalBehold refuses the shapes of an optional behold cost that
// would compile and then ask for something the card does not print:
//
//   - an optional cost keyed BeholdKey that is not a behold of one
//     creature type (no Reveal component, a reveal that is not a behold,
//     a colour quality, a repeat, or any other component beside it), or
//     an optional reveal under another key;
//   - a behold that chooses its type (#5406) built any other way than
//     OptionalBeholdChosenType: a printed type beside the choice, a
//     count below two, a quality wider than creatures, or a flash
//     opening; and a fixed-type behold with a count or the creature
//     narrowing, which OptionalBehold never writes;
//   - two of them, or one beside a mandatory reveal / behold branch:
//     a cast names one card on reveal_ids.
func checkOptionalBehold(spec Spec) {
	seen := 0
	for _, oc := range spec.OptionalCosts {
		if oc.Key != game.BeholdKey && oc.Reveal == nil {
			continue
		}
		seen++
		rest := oc
		rest.Reveal, rest.Key, rest.Label, rest.Optional, rest.Targets = nil, "", "", false, nil
		chosen := oc.Reveal != nil && oc.Reveal.ChosenType
		if oc.Key != game.BeholdKey || oc.Reveal == nil || !oc.Reveal.Behold || oc.Reveal.Exile || oc.Reveal.Color != "" ||
			(oc.Reveal.Subtype == "") != chosen || (chosen && (oc.Reveal.Count < 2 || !oc.Reveal.Creature || oc.AsThoughFlash)) ||
			(!chosen && (oc.Reveal.Count != 0 || oc.Reveal.Creature)) ||
			oc.Repeat > 1 || oc.MinX != 0 || !rest.Empty() {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q is a malformed behold cost — build it with OptionalBehold", spec.Name, oc.Key))
		}
	}
	if seen > 1 {
		panic(fmt.Sprintf("effects.Register: %q declares %d optional behold costs — a cast names one card on reveal_ids", spec.Name, seen))
	}
	if seen == 0 || spec.AdditionalCost == nil {
		return
	}
	for _, b := range append([]game.AdditionalCost{*spec.AdditionalCost}, spec.AdditionalCost.Either...) {
		if b.Reveal != nil {
			panic(fmt.Sprintf("effects.Register: %q pairs an optional behold with a mandatory reveal / behold — they share reveal_ids", spec.Name))
		}
	}
}
