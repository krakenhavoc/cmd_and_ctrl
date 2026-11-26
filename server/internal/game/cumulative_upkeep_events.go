package game

import "github.com/google/uuid"

// cumulative_upkeep_events.go — #6188: "Whenever this creature's
// cumulative upkeep is paid" (Balduvian Fallen, Hibernation's End) and
// "When a player doesn't pay this enchantment's cumulative upkeep" (Heart
// of Bogardan, Thought Lash).
//
//	702.24a Cumulative upkeep is a triggered ability that imposes an
//	        increasing cost on a permanent. "Cumulative upkeep [cost]"
//	        means "At the beginning of your upkeep, if this permanent is
//	        on the battlefield, put an age counter on this permanent.
//	        Then you may pay [cost] for each age counter on it. If you
//	        don't, sacrifice it." ...
//
// The two answers of the cumulative upkeep prompt announce themselves
// (cards/effects/cumulative_upkeep.go): the payment once it has been
// made, with the mana it spent, and the decline before the sacrifice.

// AnnounceCumulativeUpkeepPaidForEffect emits EventCumulativeUpkeepPaid
// for `cardID`, paid by `payer` for `age` age counters with `spent` (the
// mana the payment spent, none for a payment that is not mana).
//
// Caller must hold g.mu.
func (g *Game) AnnounceCumulativeUpkeepPaidForEffect(cardID, payer uuid.UUID, age int, spent []ManaToken) {
	var symbols []string
	for _, t := range spent {
		symbols = append(symbols, t.Color)
	}
	g.EmitEvent(Event{Kind: EventCumulativeUpkeepPaid, Actor: payer, Source: cardID, CardID: cardID, Amount: age, SpentSymbols: symbols})
}

// AnnounceCumulativeUpkeepNotPaidForEffect emits
// EventCumulativeUpkeepNotPaid for `cardID`, declined by `player` with
// `age` age counters on it.
//
// Caller must hold g.mu.
func (g *Game) AnnounceCumulativeUpkeepNotPaidForEffect(cardID, player uuid.UUID, age int) {
	g.EmitEvent(Event{Kind: EventCumulativeUpkeepNotPaid, Actor: player, Source: cardID, CardID: cardID, Amount: age})
}

// SpentSymbolsOf counts the mana of `colors` an
// EventCumulativeUpkeepPaid's payment spent: Balduvian Fallen's "for
// each {B} or {R} spent this way" is SpentSymbolsOf(ev, "B", "R").
func SpentSymbolsOf(ev Event, colors ...string) int {
	n := 0
	for _, s := range ev.SpentSymbols {
		for _, c := range colors {
			if s == c {
				n++
				break
			}
		}
	}
	return n
}
