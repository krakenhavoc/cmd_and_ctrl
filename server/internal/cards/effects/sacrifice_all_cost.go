package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_all_cost.go — #2097: "As an additional cost to cast this
// spell, sacrifice all <permanents> you control" (Soulblast). The
// constructor and its Register guard. The engine half, which decides
// WHICH permanents go, is game/sacrifice_all_cost.go.

// SacrificeAllCost is "As an additional cost to cast this spell,
// sacrifice all <permanents> you control": SacrificeAllCost("creatures
// you control", Creature()) is Soulblast's.
//
// The caster chooses nothing. The engine takes every permanent they
// control that matches, as the spell is cast (CR 601.2b, 601.2h), and
// none is a legal payment: an empty board pays the cost in full (CR
// 118.3 asks only for what the cost needs, and it needs nothing then).
// A phased-out permanent is not taken (CR 702.26b); an indestructible
// one is (CR 701.21a). They leave as one simultaneous exit with the
// spell on the stack, and the record is the usual one: ctx.Sacrificed()
// counts them and ctx.SacrificedPermanents() / SacrificedTotalPower()
// read each as it last existed on the battlefield (ADR 0113 §1).
//
// The clause carries the "any number" bounds, so every reader that
// asks how many a payment may name (the validator, the view, the
// enumerator) already allows the whole board and zero; the flag on the
// cost says that the number is "all of them". Only the mandatory slot
// may carry it (checkSacrificeAllCost).
func SacrificeAllCost(label string, preds ...CardPredicate) *game.AdditionalCost {
	return &game.AdditionalCost{
		Sacrifice:    sacrificeSpec(label, preds...).WithCount(0, 0),
		SacrificeAll: true,
		Label:        "Sacrifice all " + label,
	}
}

// checkSacrificeAllCost is Register's guard for SacrificeAll. Each
// refused shape compiles and then pays something no card prints:
//
//   - the flag with no sacrifice clause, or on a clause that is not the
//     open "any number" shape (a fixed or X count would be checked
//     against a set the caster did not choose);
//   - a set rule on the clause ("a Swamp and a Forest" is a choice);
//   - the flag on an optional cost or an either/or branch: every printed
//     card has it as the one mandatory cost, and a branch or a kicker
//     that took the whole board would need its own announcement.
//
// checkVariableSacrificePlan already keeps any other sacrifice out of
// the same plan, because the clause is a variable one.
func checkSacrificeAllCost(spec Spec) {
	for _, oc := range spec.OptionalCosts {
		if oc.SacrificeAll {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q sacrifices all — only a cast's mandatory additional cost prints that (SacrificeAllCost, #2097)", spec.Name, oc.Key))
		}
	}
	ac := spec.AdditionalCost
	if ac == nil {
		return
	}
	for i, b := range ac.Either {
		if b.SacrificeAll {
			panic(fmt.Sprintf("effects.Register: %q either/or branch %d sacrifices all — only a cast's mandatory additional cost prints that (SacrificeAllCost, #2097)", spec.Name, i))
		}
	}
	if !ac.SacrificeAll {
		return
	}
	switch {
	case ac.Sacrifice == nil:
		panic(fmt.Sprintf("effects.Register: %q sets SacrificeAll with no sacrifice clause — build it with SacrificeAllCost", spec.Name))
	case !game.SacrificeAnyNumber(ac.Sacrifice):
		panic(fmt.Sprintf("effects.Register: %q sets SacrificeAll on a clause with a count — build it with SacrificeAllCost, whose bounds are \"any number\" (#2097)", spec.Name))
	case len(ac.Sacrifice.EachOf) > 0:
		panic(fmt.Sprintf("effects.Register: %q sets SacrificeAll on a clause with a set rule — \"all\" chooses nothing (#2097)", spec.Name))
	}
}
