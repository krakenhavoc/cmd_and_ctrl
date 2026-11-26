package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// escape_counter_choice.go — "This creature escapes with your choice of a
// +1/+1 counter or a flying counter on it" (Tizerus Charger, #6154,
// CR 702.138c).
//
// CR 702.138c reads "[this permanent] escapes with [counters]" as "if this
// permanent escaped, it enters with [those counters]". EscapeWithCounters
// fixes the kind on the escape cost; a choice of kinds is instead the
// permanent's AsEnters hook, gated on the permanent's own escape record
// (Card.Escaped, stamped before the hook runs), asking the same "enters
// with your choice of counters" question S82OldB113EntersWithChoiceOfCounters
// asks for a printed entry choice. So the card keeps its one plain Escape
// offer (and its "escaped" provenance), the choice is made as it arrives,
// before anything can act, and a creature cast from hand, reanimated or
// blinked is asked nothing.

// EscapesWithChoiceOfCounters is the AsEnters hook for "escapes with your
// choice of <A> or <B> counter on it": each choice is the counter kinds one
// answer puts on the creature, in printed order. It asks only when the
// permanent escaped.
func EscapesWithChoiceOfCounters(name string, choices ...[]string) func(*game.Card, *Context) error {
	ask := S82OldB113EntersWithChoiceOfCounters(name, choices...)
	return func(card *game.Card, ctx *Context) error {
		if card == nil || !card.Escaped() {
			return nil
		}
		return ask(card, ctx)
	}
}
