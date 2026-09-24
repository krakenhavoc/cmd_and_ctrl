package roadmap

import (
	"regexp"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// probes.go — the vocabulary registry.go's probes are written in.
//
// Every Spec probe here reads a DECLARATION: a keyword in
// PrintedKeywords, a cost under a known key, a special action of a
// known kind. None of them inspects a closure, because a guess about
// what a closure does is not something a public page should repeat.
// Mechanics with no declaration to read use a printed-text pattern
// instead (Item.Printed), which asks the card rather than the code.

// printedKeyword matches a printed keyword line: the keyword at the
// start of a line, alone or in a comma-separated keyword list
// ("Flying, vigilance"). It is the printed-text twin of
// printedKeywordProbe, for cards that print the keyword without a
// catalog declaration of it.
func printedKeyword(kw string) string {
	return `(?im)^(?:[a-z' -]+, )*` + regexp.QuoteMeta(kw) + `\b`
}

// printedLine matches a line that begins with the given words — a
// keyword with a cost or a number after it ("Kicker {2}", "Ward {2}",
// "Hideaway 4").
func printedLine(words string) string {
	return `(?im)^` + regexp.QuoteMeta(words) + `\b`
}

// printedWords matches the words anywhere in the printed text.
func printedWords(words string) string {
	return `(?i)\b` + regexp.QuoteMeta(words) + `\b`
}

// hasKeyword reports whether the spec declares one of the canonical
// keyword tokens in PrintedKeywords. A token ending in a space is a
// prefix ("protection from ").
func hasKeyword(tokens ...string) func(effects.Spec) bool {
	return func(s effects.Spec) bool {
		for _, kw := range s.PrintedKeywords {
			kw = strings.ToLower(kw)
			for _, t := range tokens {
				if kw == t || (strings.HasSuffix(t, " ") && strings.HasPrefix(kw, t)) {
					return true
				}
			}
		}
		return false
	}
}

// altCost reports whether the card offers an alternative cost under
// key, through the engine's own lookup (the probe coverage.altCost
// uses).
func altCost(keys ...string) func(effects.Spec) bool {
	return func(s effects.Spec) bool {
		for _, k := range keys {
			if game.AlternativeCostByKey(s.OracleID, k) != nil {
				return true
			}
		}
		return false
	}
}

// optionalCost reports whether the card declares an optional
// additional cost under key (kicker, multikicker, buyback, offspring).
func optionalCost(key string) func(effects.Spec) bool {
	return func(s effects.Spec) bool {
		for _, oc := range s.OptionalCosts {
			if oc.Key == key {
				return true
			}
		}
		return false
	}
}

// activated reports whether any activated ability satisfies pred.
func activated(pred func(effects.ActivatedAbility) bool) func(effects.Spec) bool {
	return func(s effects.Spec) bool {
		for _, ab := range s.Activated {
			if pred(ab) {
				return true
			}
		}
		return false
	}
}

// activatedLabel matches an activated ability whose Label begins with
// prefix. Only used for keyword constructors that write their own
// label ("Embalm {cost} (…)"), so the label is the constructor's
// signature rather than prose a card author chose.
func activatedLabel(prefix string) func(effects.Spec) bool {
	return activated(func(ab effects.ActivatedAbility) bool { return strings.HasPrefix(ab.Label, prefix) })
}

// specialAction reports whether the card declares a CR 116.2 special
// action of the given kind.
func specialAction(kind game.SpecialActionKind) func(effects.Spec) bool {
	return func(s effects.Spec) bool {
		for _, sa := range s.SpecialActions {
			if sa.Kind == kind {
				return true
			}
		}
		return false
	}
}

// tapCost reports whether the spell's tap-to-pay cost has the key
// (convoke, waterbend).
func tapCost(key string) func(effects.Spec) bool {
	return func(s effects.Spec) bool {
		if s.TapCost != nil && s.TapCost.Key == key {
			return true
		}
		for _, ab := range s.Activated {
			if ab.Cost.Waterbend != nil && ab.Cost.Waterbend.Key == key {
				return true
			}
		}
		return false
	}
}

// designation reports whether any printed ability is gated on one of
// the designation kinds (a Class level, a solved Case, charge
// counters, harnessed).
func designation(kinds ...game.DesignationKind) func(effects.Spec) bool {
	want := func(d game.Designation) bool {
		for _, k := range kinds {
			if d.Kind == k {
				return true
			}
		}
		return false
	}
	return func(s effects.Spec) bool {
		for _, a := range s.Static {
			if want(a.ActiveWhen) {
				return true
			}
		}
		for _, a := range s.Triggered {
			if want(a.ActiveWhen) {
				return true
			}
		}
		for _, a := range s.Activated {
			if want(a.ActiveWhen) {
				return true
			}
		}
		for _, a := range s.CostModifiers {
			if want(a.ActiveWhen) {
				return true
			}
		}
		return false
	}
}

// anyOf is true when any probe is.
func anyOf(ps ...func(effects.Spec) bool) func(effects.Spec) bool {
	return func(s effects.Spec) bool {
		for _, p := range ps {
			if p(s) {
				return true
			}
		}
		return false
	}
}
