package game

import "github.com/google/uuid"

// tribute.go — tribute (CR 702.104, #6251).
//
//	702.104a Tribute is a static ability that functions as the creature
//	         with tribute is entering the battlefield. "Tribute N" means
//	         "As this creature enters, choose an opponent. That player
//	         may put an additional N +1/+1 counters on it as it enters."
//	702.104b Objects with tribute have triggered abilities that check
//	         "if tribute wasn't paid." This condition is true if the
//	         opponent chosen as a result of the tribute ability didn't
//	         have the creature enter the battlefield with +1/+1 counters
//	         as specified by the creature's tribute ability.
//
// Two entry replacements on the creature itself (CR 614.12), built by
// effects.Tribute and asked before it enters (CR 614.12a):
//
//   - "choose an opponent": an EntryController effect whose purpose is
//     ControlForTribute. It asks the entry_controller question Captive
//     Audience asks, with the same seat buttons and the same defaults (one
//     opponent is chosen without asking, an entry that cannot pause takes
//     the first opponent after the chooser), but the answer is recorded on
//     ReplacementEvent.TributeOpponent and the control does not change.
//   - "that player may put N +1/+1 counters on it": an Optional
//     replacement that applies only once TributeOpponent is set, whose
//     Controller is that opponent, so the existing "may" prompt is put to
//     them. Its Replace adds the counters and sets EntersTributePaid,
//     which the landing copies onto Card.TributePaid.
//
// A declined, unaskable or unanswerable payment leaves TributePaid false,
// so "if tribute wasn't paid" is true: the opponent "didn't have the
// creature enter with the counters" (CR 702.104b).

// setTributeOpponentLocked records the opponent chosen for a tribute.
//
// Caller must hold g.mu.
func (g *Game) setTributeOpponentLocked(ev *ReplacementEvent, chooser, opponent uuid.UUID) {
	if ev == nil || opponent == uuid.Nil {
		return
	}
	ev.TributeOpponent = opponent
	g.EmitEvent(Event{
		Kind:   EventPlayerChosen,
		Actor:  chooser,
		CardID: ev.CardID,
		Target: opponent,
		Label:  g.seatLabelLocked(opponent),
	})
}

// applyEntersTributePaidLocked is the battlefield landing's half of
// ReplacementEvent.EntersTributePaid.
//
// Caller must hold g.mu in write mode.
func (g *Game) applyEntersTributePaidLocked(cardID uuid.UUID, paid bool) {
	if !paid {
		return
	}
	if c := findBattlefieldCard(g, cardID); c != nil {
		c.TributePaid = true
	}
}

// TributePaidFor reports whether the tribute of the permanent `id` was
// paid as it entered (CR 702.104b). False for a permanent that is not
// on the battlefield; a trigger whose source may have left reads
// PermanentInfo.TributePaid, its last-known information.
//
// Caller must hold either lock.
func (g *Game) TributePaidFor(id uuid.UUID) bool {
	if c := findBattlefieldCard(g, id); c != nil {
		return c.TributePaid
	}
	return false
}
