package game

import "github.com/google/uuid"

// activate_only_once.go — "Activate only once." on an activated ability
// that is not an exhaust ability (#5221). Possessed Goat's "{3},
// Discard a card: Put three +1/+1 counters on this creature … Activate
// only once.", Haunted Screen's "{7}: … Activate only once."
//
// The printed text is an activation instruction (CR 602.1b), and it
// limits the ability of THAT object: CR 602.5b, "If an activated
// ability has a restriction on its use (for example, 'Activate only
// once each turn'), the restriction continues to apply to that object
// even if its controller changes." So it is the exhaust limit, word
// for word — CR 702.177a defines exhaust as "[Cost]: [Effect]. Activate
// only once." — without the keyword.
//
// # Why it is not Exhaust: true
//
// The keyword is more than the limit. "Whenever you activate an exhaust
// ability" (Rangers' Refueler) watches Event.Exhaust, "Exhaust abilities
// … cost {2} less" (Boom Scholar) reads AbilityCostSubject.Exhaust, and
// "You may activate exhaust abilities as though they haven't been
// activated" (Elvish Refueler) suspends the exhaust gate. None of them
// reaches an ability that merely says "Activate only once", so this is
// its own bit, read in its own gate, and it is never stamped on an event
// or a cost query.
//
// # What it shares with exhaust
//
// The record: Game.Activations.Ever, keyed by (object, label) and
// written at the announce on the one path that pays for an activation
// (activation_tally.go). Everything exhaust promises falls out of that
// key the same way:
//
//   - a countered or fizzled activation has still been activated;
//   - a new object after a zone change (CR 400.7) reads a key nothing
//     has written, so a flickered Possessed Goat may activate it again;
//   - a change of control is not a zone change, so the ability stays
//     spent for its new controller (CR 602.5b);
//   - the record is in the snapshot, so a restore keeps it spent.
//
// # One gate, three readers
//
// OnlyOnceSpent is the whole rule, for the reason AbilityExhausted is
// one function: ActivateCatalogAbility refuses with
// ErrAbilityActivatedOnce before anything is paid, internal/legal does
// not offer the move (#544), and protocol's ActivatedAbilityView stamps
// `once_used` so the client greys the row.

// OnlyOnceSpent reports whether `ab` says "Activate only once" and the
// object `source` names has already activated it. False for every
// ability without the bit, which is the first test, so the ordinary
// ability costs one bool read.
//
// No asker: no permission suspends this limit (Elvish Refueler's names
// exhaust abilities), so the answer is the same for every player.
//
// Caller must hold g.mu.
func (g *Game) OnlyOnceSpent(source uuid.UUID, ab ActivatedAbilityShape) bool {
	if !ab.OnlyOnce {
		return false
	}
	return g.ActivatedThisGame(source, ab.Label) > 0
}
