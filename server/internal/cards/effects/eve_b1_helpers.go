package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// eve_b1_helpers.go — the shapes Eventide, batch 1 (#5590) prints more
// than once: the "whenever you cast a <colour> spell, you may untap this
// creature" cycle (Nettle Sentinel, Cinder Pyromancer, Ballynock Trapper,
// Merrow Levitator) and the Hatchlings' four -1/-1 counters that each
// spell of their colours removes one at a time.

// eveB1UntapOnCast is "Whenever you cast a <colour> spell, you may untap
// this creature." The colour is read off the spell as it is cast, so a
// multicoloured spell of that colour counts, and the "may" is asked as
// the trigger resolves (CR 603.5). A source no longer on the battlefield
// untaps nothing.
func EveB1UntapOnCast(name, colour string) game.TriggeredAbility {
	return Optional(WheneverYouCast(OfColor(colour), name+" — untap it", B31UntapSelf),
		name+" — untap this creature?")
}

// eveB1HatchlingCounters is the Hatchlings' "This creature enters with four
// -1/-1 counters on it." (CR 614.1c).
func EveB1HatchlingCounters(name string) game.ReplacementEffect {
	return B10EntersWithCounters(game.CounterMinusOne, 4, name+": enters with four -1/-1 counters")
}

// eveB1RemoveCounterOnCast is "Whenever you cast a <colour> spell, remove a
// -1/-1 counter from this creature." One trigger per printed colour, each
// with its own stack label, so a spell of both colours puts both on the
// stack. With no -1/-1 counter left it removes nothing.
func EveB1RemoveCounterOnCast(name, colour string) game.TriggeredAbility {
	return WheneverYouCast(OfColor(colour),
		name+" — "+game.ColorName(colour)+" spell: remove a -1/-1 counter from this creature",
		Do(Instruction(StdB64RemoveAMinusCounterFromThis)))
}
