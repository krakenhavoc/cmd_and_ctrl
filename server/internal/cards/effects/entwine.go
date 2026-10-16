package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entwine.go — CR 702.42a: "'Entwine [cost]' means 'You may choose all
// modes of this spell instead of just the number specified. If you do,
// you pay an additional [cost].' Using the entwine ability follows the
// rules for choosing modes and paying additional costs in rules 601.2b
// and 601.2f–h." (#3866, ADR 0065's 2026-10-16 amendment)
//
// Entwine is two things the engine already has, joined by the
// announcement CR 601.2b makes them share:
//
//   - the cost is an ADR 0073 optional additional cost, keyed
//     game.EntwineKey and paid at most once. Its mana joins the total at
//     CR 601.2f beside every other additional cost, a sacrifice rides
//     sacrifice_ids and is paid with the spell on the stack, and the
//     client offers it where it offers kicker;
//   - the mode count is #1655's announcement-driven count: with the
//     entwine cost announced the spell chooses every bullet, no more and
//     no fewer (InsteadIf, so the minimum rises with the maximum), and
//     without it the printed count stands.
//
// So the cast gate, the hand card's `modes.if_optional_paid`, the
// picker and the bot's enumerator all read one game.ModeBoundsForEffect,
// and none of them can offer a selection another refuses. CR 702.42b's
// "follow the text of each of the modes in the order written" is the
// printed-order walk every modal spell already takes (PrintedModeOrder).
//
//	OptionalCosts: []game.AdditionalCost{Entwine("{2}")},
//	Modes:         Entwining(ChooseOne(…)),                  // Spectral Shift

// Entwine is "Entwine {cost}".
func Entwine(mana string) game.AdditionalCost {
	return game.AdditionalCost{
		Optional: true,
		Key:      game.EntwineKey,
		ManaCost: mana,
		Label:    "Entwine " + mana,
	}
}

// EntwineSacrifice is a non-mana entwine cost, "Entwine—Sacrifice N
// <permanents>" (Betrayal of Flesh's three lands, Solar Tide's two):
// the same counted sacrifice clause SacrificeNCost builds, so the
// caster names exactly n and they leave together with the spell on the
// stack.
func EntwineSacrifice(n int, label string, preds ...CardPredicate) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:  true,
		Key:       game.EntwineKey,
		Sacrifice: sacrificeSpec(label, preds...).WithCount(n, n),
		Label:     "Entwine—Sacrifice " + label,
	}
}

// WasEntwined is the mode count's condition: the entwine cost was
// announced with the modes (CR 601.2b). Registered through
// ModeConditionOnAnnouncement because it reads the announcement, not
// the board — the same shape as WasKicked.
var WasEntwined = game.ModeConditionOnAnnouncement("was-entwined", func(_ *game.Game, q game.ModeCountQuery) bool {
	return q.OptionalCostTimes(game.EntwineKey) > 0
})

// Entwining raises a modal spell's count to every bullet while its
// entwine cost is announced: "Choose one — … Entwine {2}" is
// Entwining(ChooseOne(…)) beside OptionalCosts: Entwine("{2}"). Kaya's
// Guile's "Choose two … Entwine {3} (Choose all if you pay the entwine
// cost.)" is Entwining(ChooseN("Choose two", 2, 2, …)): two, or all
// four, never three.
func Entwining(ms *game.ModeSpec) *game.ModeSpec {
	return ms.InsteadIf(len(ms.Options), WasEntwined)
}

// checkEntwine holds the two halves together: a card with an entwine
// cost raises its spell's mode count with Entwining, and a mode count
// raised by WasEntwined belongs to a card with an entwine cost — either
// half alone is a cost that buys nothing, or a count that can never
// rise. Entwine is a static ability of a modal SPELL (CR 702.42a), so a
// trigger's or an activated ability's modes never read it.
func checkEntwine(spec Spec) {
	entwined := func(ms *game.ModeSpec) bool {
		return ms != nil && ms.RaiseMaxIf.Key() == WasEntwined.Key()
	}
	for _, a := range spec.Activated {
		if entwined(a.Modes) {
			panic(fmt.Sprintf("effects.Register: %q raises an activated ability's modes on entwine — CR 702.42a is a static ability of a modal spell", spec.Name))
		}
	}
	for _, t := range spec.Triggered {
		if entwined(t.Modes) {
			panic(fmt.Sprintf("effects.Register: %q raises a trigger's modes on entwine — CR 702.42a is a static ability of a modal spell", spec.Name))
		}
	}
	n := 0
	for _, oc := range spec.OptionalCosts {
		if oc.Key != game.EntwineKey {
			continue
		}
		n++
		if oc.MaxPayments() != 1 {
			panic(fmt.Sprintf("effects.Register: %q declares an entwine cost paid more than once — CR 702.42a chooses every mode once", spec.Name))
		}
		if oc.ManaCost == "" && oc.Sacrifice == nil {
			panic(fmt.Sprintf("effects.Register: %q declares an entwine cost that demands nothing", spec.Name))
		}
	}
	ms := spec.Modes
	switch {
	case n == 0 && entwined(ms):
		panic(fmt.Sprintf("effects.Register: %q raises its modes on entwine but declares no entwine cost — add Entwine(…) to OptionalCosts", spec.Name))
	case n == 0:
		return
	case n > 1:
		panic(fmt.Sprintf("effects.Register: %q declares %d entwine costs — CR 702.42a prints one", spec.Name, n))
	case !entwined(ms):
		panic(fmt.Sprintf("effects.Register: %q declares an entwine cost but its modes do not rise with it — build them with Entwining(…)", spec.Name))
	case ms.Repeatable:
		panic(fmt.Sprintf("effects.Register: %q declares entwine on a spell that may repeat a mode — \"all modes\" has no count there", spec.Name))
	case ms.Escalate != nil:
		panic(fmt.Sprintf("effects.Register: %q declares both entwine and escalate — no printed card has both, and the two costs would price one count twice", spec.Name))
	case ms.RaisedMin != len(ms.Options) || ms.RaisedMax != len(ms.Options):
		panic(fmt.Sprintf("effects.Register: %q raises its modes on entwine to %d..%d, not all %d — build them with Entwining(…)", spec.Name, ms.RaisedMin, ms.RaisedMax, len(ms.Options)))
	}
}
