package game

import "github.com/google/uuid"

// instant_window.go — #1600: "Activate only as an instant" on a MANA
// ability (Lion's Eye Diamond, Diamond Lion).
//
// A mana ability has its own window (CR 605.3a): whenever its
// controller has priority, AND whenever a payment is being made — in
// the middle of casting a spell, inside another ability's cost, in the
// middle of a resolution that asks for mana. That second half is why
// ActivationTimingOpenLocked returns true for a mana ability before it
// walks anything, and why the dispatcher does not ask the activator to
// hold priority for activate_mana_ability.
//
// "Activate only as an instant" takes the second half away. CR 602.5e:
// the player "must follow the timing rules for casting an instant
// spell", and CR 117.1a's timing rule is "any time they have
// priority" — which is never in the middle of casting a spell or of a
// resolution. The 2004-10-04
// ruling on Lion's Eye Diamond says it in as many words — "it can only
// be activated at times when you can cast an instant" — and Diamond
// Lion's 2021-06-18 ruling draws the consequence: "you can't activate
// the ability intending to use the mana to cast a spell from your
// hand". It is still a mana ability: it does not use the stack and
// cannot be responded to (CR 605.3b).
//
// The engine has no window in which a player clicks mana abilities
// while a spell is half-cast — a cast arrives whole, with its payment
// named, and anything it pays automatically goes through the
// auto-tapper. So "never mid-cast" is two facts:
//
//   - the AUTO-TAPPER never plans the ability. A "Discard your hand"
//     cost already keeps it out (autoTapAbilityAccepts refuses every
//     discard component), and that is the only shape that prints the
//     restriction; the planner's refusal is pinned for it by name.
//   - a hand-click is refused unless this window is open. Every pause
//     in which the engine waits on a player — a pay-unless ("counter
//     it unless its controller pays {3}"), a mana_pick, a CR 903.9
//     answer that parked a cast, any blocking prompt — is a point in
//     the middle of a resolution or an announcement, which is exactly
//     where a mana ability may otherwise be activated and this one may
//     not.
//
// The predicate is the legal enumerator's own reading of "this seat
// may act now" (internal/legal enumerateLocked: holds priority, owes
// no prompt, no blocking prompt is open, the mulligan window is shut),
// so the bot is offered the activation exactly when the engine
// accepts it (#544). It is read through a mana ability's Condition
// (effects.OnlyAsAnInstant), so the click path, the enumerator, the
// view's condition_unmet and the auto-tapper's executor ask one
// question.
//
// Split second (CR 702.61b) does not close it: split second stops
// abilities that AREN'T mana abilities, and this one is.

// InstantWindowOpenForEffect reports whether `player` could cast an
// instant right now: the game is running, the mulligan window is shut,
// `player` holds priority, owes no prompt, and no prompt that stops
// the table is open.
//
// CALLER MUST ALREADY HOLD g.mu (read or write): an activation
// Condition runs under it on every path that asks one.
func (g *Game) InstantWindowOpenForEffect(player uuid.UUID) bool {
	if g == nil || player == uuid.Nil || g.State != StateActive || g.MulligansOpen {
		return false
	}
	ph := g.Turn.PriorityHolder
	if ph < 0 || ph >= len(g.Seats) || g.Seats[ph] == nil || g.Seats[ph].ID != player {
		return false
	}
	for _, c := range g.PendingChoices {
		if c == nil {
			continue
		}
		// A prompt addressed to this player is a question they are in
		// the middle of answering — a pay-unless tax is the classic
		// one, and the reason the restriction exists.
		if c.Chooser == player {
			return false
		}
		// Somebody else's prompt that stops the table is a resolution
		// still under way (choice_gate.go).
		if g.ChoicePromptBlocksTable(c) {
			return false
		}
	}
	return true
}
