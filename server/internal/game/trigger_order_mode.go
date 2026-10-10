package game

import "fmt"

// TriggerOrderMode is a seat's answer to "when should the game ask me to
// order my triggers?" (#1968, ADR 0018's #1968 amendment). It replaces
// #1530's always-ask boolean with three values. The CR 603.3b order is
// always the player's choice; the mode says only when the game asks for
// it and when it takes the order the batch was collected in instead.
type TriggerOrderMode string

const (
	// TriggerOrderWhenItMatters asks only when the order can change the
	// game: seatNeedsTriggerOrder's skips apply. The default, and the
	// zero value, so a seat that never chose has it.
	TriggerOrderWhenItMatters TriggerOrderMode = ""
	// TriggerOrderAlways asks for every batch of two or more not yet
	// ordered, the skips included (#1530's "always ask").
	TriggerOrderAlways TriggerOrderMode = "always"
	// TriggerOrderNever never asks: the seat's batch goes on the stack
	// in the order its triggers were collected, which is the order an
	// auto-ordered batch has always used (#1511).
	TriggerOrderNever TriggerOrderMode = "never"
)

// TriggerOrderWireWhenItMatters is the default mode's wire spelling.
// Inside the engine the default is the zero value; on the wire and in
// the client's settings it has a name.
const TriggerOrderWireWhenItMatters = "when_it_matters"

// ParseTriggerOrderMode reads a mode from its wire spelling:
// "when_it_matters", "always" or "never". Anything else is an error.
func ParseTriggerOrderMode(s string) (TriggerOrderMode, error) {
	switch s {
	case TriggerOrderWireWhenItMatters:
		return TriggerOrderWhenItMatters, nil
	case string(TriggerOrderAlways):
		return TriggerOrderAlways, nil
	case string(TriggerOrderNever):
		return TriggerOrderNever, nil
	}
	return "", fmt.Errorf("unknown trigger order mode %q", s)
}

// Wire is the mode's wire spelling, the inverse of ParseTriggerOrderMode.
func (m TriggerOrderMode) Wire() string {
	if m == TriggerOrderWhenItMatters {
		return TriggerOrderWireWhenItMatters
	}
	return string(m)
}

// valid reports whether m is one of the three modes.
func (m TriggerOrderMode) valid() bool {
	switch m {
	case TriggerOrderWhenItMatters, TriggerOrderAlways, TriggerOrderNever:
		return true
	}
	return false
}

// restoredTriggerOrder is the mode a snapshot seat restores with. A file
// from #1968 on names the mode (triggerOrder); a file from before it has
// only #1530's boolean, where true meant "always". A mode this binary
// does not know (only a newer binary could have written one) is kept as
// it was, so a roll-forward gets it back, and until then it behaves as
// the default: seatNeedsTriggerOrder skips only for "never" and asks
// for everything only for "always", so an unknown mode asks only when
// the order matters, which never takes a choice away the default would
// offer.
func restoredTriggerOrder(mode string, alwaysAsk bool) TriggerOrderMode {
	if mode != "" {
		return TriggerOrderMode(mode)
	}
	if alwaysAsk {
		return TriggerOrderAlways
	}
	return TriggerOrderWhenItMatters
}
