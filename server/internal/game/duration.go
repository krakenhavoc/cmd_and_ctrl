package game

import "github.com/google/uuid"

// duration.go is the CR 611.2 duration model (ADR 0063, #755).
//
// A continuous effect created by a resolved spell or ability lasts for
// a stated duration, and before S38 the engine knew exactly one of
// them: "until end of turn". `ScopedStatic.ExpiresAfterTurn` was an
// int stamped with the round number and swept at the first cleanup
// step, which is right for Giant Growth and cannot express Agent of
// Treachery ("gain control of target permanent", no duration at all),
// Sower of Temptation ("for as long as this creature remains on the
// battlefield") or Mass Diminish ("until your next turn").
//
// `Duration` is the replacement, and it is DATA — no closures, no
// pointers into game state. That matters three times over:
//
//   - `Clone` copies a `ScopedStatic` by value, so a duration that
//     captured a `*Card` would alias across an undo snapshot.
//   - The snapshot census already refuses a restore point for a game
//     holding a scoped static, because the ability is two closures
//     (snapshot.go). The duration is the half of that record a future
//     persistence pass (#515) can write to disk verbatim.
//   - One `switch` in `durationExpiredLocked` decides when every
//     continuous effect in the game ends. A predicate closure would
//     scatter that rule across the catalog and let the copies drift.
//
// ONE EXPIRY FUNCTION, THREE MOMENTS. `durationExpiredLocked` is the
// only code that answers "is this effect over". It is reached from
// `sweepScopedStaticsLocked`, which runs at the cleanup step
// (CR 514.2), when a turn begins (the "until your next turn"
// boundary), and at the top of every layer recompute (where a
// "for as long as" condition can have gone false). Adding a duration
// kind is a new case in that switch and nothing else.

// DurationKind names the four shapes CR 611.2 gives a continuous
// effect created by a resolved spell or ability.
type DurationKind int

const (
	// UntilEndOfTurn ends during the cleanup step of the turn the
	// effect was created in (CR 514.2) — including an effect created
	// during that turn's END step, which is not the end of the turn.
	// The zero value, so an unstamped duration behaves exactly as
	// the pre-S38 registry did.
	UntilEndOfTurn DurationKind = iota

	// UntilYourNextTurn ends as the named player's next turn begins
	// (CR 500.1: a turn begins with its untap step; CR 502.1 puts
	// the untapping inside that step, so the turn has begun before
	// anything untaps). Mass Diminish, Teferi's Protection.
	UntilYourNextTurn

	// ForAsLongAs ends when its Condition stops being true
	// (CR 611.2b). Sower of Temptation's "for as long as this
	// creature remains on the battlefield".
	ForAsLongAs

	// Indefinite never ends on its own: CR 611.2a, an effect with no
	// stated duration lasts until the game ends. Agent of Treachery.
	// It can still be dropped by the pin (see Duration.Pinned).
	Indefinite
)

// DurationCondition is the re-evaluated half of a ForAsLongAs
// duration. A small closed vocabulary rather than a predicate,
// because a predicate would be a closure in snapshotted state.
type DurationCondition int

const (
	// WhileSourceOnBattlefield — "for as long as ~ remains on the
	// battlefield". The source must still be the SAME object:
	// CR 400.7 makes a permanent that left and returned a new one,
	// so both the instance ID and the battlefield-entry stamp have
	// to match. A Sower of Temptation blinked in response gives the
	// creature back.
	WhileSourceOnBattlefield DurationCondition = iota

	// WhileYouControlSource — "for as long as you control ~". The
	// above, and Duration.Player must still be the source's
	// controller.
	WhileYouControlSource
)

// Duration is how long one continuous effect lasts. The zero value is
// "until end of turn" with no stamp, which the sweep treats as ending
// at the first cleanup it sees — the pre-S38 behaviour.
//
// IMMUTABLE after registration, like every other field on
// ScopedStatic: `Clone` shares the value with every undo snapshot.
type Duration struct {
	// Kind selects which of the fields below mean anything.
	Kind DurationKind

	// Player is the player the duration is ABOUT: whose next turn
	// ends it (UntilYourNextTurn), or who has to keep controlling
	// the source (ForAsLongAs + WhileYouControlSource).
	Player uuid.UUID

	// ExpiresAtTurnsBegun is the value of Player.TurnsBegun at which
	// an UntilYourNextTurn duration is over — stamped as
	// `TurnsBegun + 1` when the effect is created, so it names that
	// player's next turn whether the effect was made on their turn
	// or on somebody else's.
	//
	// Counting seat-turns rather than reading Turn.Number is the
	// whole point: Turn.Number counts ROUNDS, so all four seats in a
	// Commander game share one number and "your next turn" cannot be
	// expressed with it (ADR 0035 §3's caveat, retired by ADR 0063
	// Decision 3). Player.TurnsBegun also goes up for a seat the
	// rotation STEPS OVER because that player has left, which is
	// CR 800.4m: the effect ends when their turn would have begun.
	ExpiresAtTurnsBegun int

	// ExpiresAfterTurnsBegun is the same counter for the creating
	// player, stamped at registration for an UntilEndOfTurn effect.
	// The cleanup sweep does not need it — "end of turn" is whatever
	// cleanup step comes next — but a turn that ends EARLY (its
	// active player left; ADR 0059 Decision 6) can run the cleanup
	// sweep from an arbitrary step, and this is the backstop that
	// guarantees the grant cannot outlive the turn it was made in.
	ExpiresAfterTurnsBegun int

	// Condition is which ForAsLongAs test to re-run. Ignored for
	// every other kind.
	Condition DurationCondition

	// Source and SourceEnteredAt name the object a ForAsLongAs
	// duration watches, keyed the way every CR 611.2c affected set
	// in the engine is keyed: instance ID plus the battlefield-entry
	// stamp, so CR 400.7 is checked rather than assumed.
	Source          uuid.UUID
	SourceEnteredAt int64

	// Pinned and PinnedEnteredAt optionally name the ONE object the
	// effect exists to affect — the stolen permanent, for a control
	// change. When set, the effect ends as soon as that object stops
	// being on the battlefield as the same object, whatever the kind
	// says.
	//
	// This is garbage collection with a rules justification: an
	// effect that moves one permanent has nothing left to do once
	// that permanent is gone, and CR 400.7 says a returning one is
	// not it. Without the pin an Indefinite entry would sit in the
	// registry — and in the snapshot census, keeping the game off
	// the full-restore path — for the rest of the game.
	Pinned          uuid.UUID
	PinnedEnteredAt int64
}

// UntilEndOfTurnDuration is "until end of turn" (CR 514.2), stamped
// against the current turn. Caller must hold g.mu.
func (g *Game) UntilEndOfTurnDuration() Duration {
	active := g.activePlayerIDLocked()
	return Duration{
		Kind:                   UntilEndOfTurn,
		Player:                 active,
		ExpiresAfterTurnsBegun: g.turnsBegunForLocked(active),
	}
}

// UntilYourNextTurnDuration is "until <player>'s next turn"
// (CR 611.2b), stamped against that player's seat-turn count. Caller
// must hold g.mu.
func (g *Game) UntilYourNextTurnDuration(player uuid.UUID) Duration {
	return Duration{
		Kind:                UntilYourNextTurn,
		Player:              player,
		ExpiresAtTurnsBegun: g.turnsBegunForLocked(player) + 1,
	}
}

// ForAsLongAsOnBattlefieldDuration is "for as long as ~ remains on the
// battlefield" (CR 611.2b). Returns the duration and false when the
// source is not on the battlefield at all: CR 611.2b says an effect
// whose condition is already false as it would start never starts, so
// the caller registers nothing.
//
// Caller must hold g.mu.
func (g *Game) ForAsLongAsOnBattlefieldDuration(source uuid.UUID) (Duration, bool) {
	c, ok := g.battlefieldCardLocked(source)
	if !ok {
		return Duration{}, false
	}
	return Duration{
		Kind:            ForAsLongAs,
		Condition:       WhileSourceOnBattlefield,
		Source:          source,
		SourceEnteredAt: c.EnteredBattlefieldAt,
	}, true
}

// ForAsLongAsYouControlDuration is "for as long as you control ~"
// (CR 611.2b). Same "never starts" rule as its sibling, plus the
// control test. Caller must hold g.mu.
func (g *Game) ForAsLongAsYouControlDuration(source, player uuid.UUID) (Duration, bool) {
	c, ok := g.battlefieldCardLocked(source)
	if !ok || c.Controller != player {
		return Duration{}, false
	}
	return Duration{
		Kind:            ForAsLongAs,
		Condition:       WhileYouControlSource,
		Player:          player,
		Source:          source,
		SourceEnteredAt: c.EnteredBattlefieldAt,
	}, true
}

// IndefiniteDuration is "no stated duration" — CR 611.2a, the effect
// lasts until the game ends. Needs no game state, so it is a plain
// function.
func IndefiniteDuration() Duration { return Duration{Kind: Indefinite} }

// PinnedTo returns a copy of d that also ends when `object` stops
// being the same object on the battlefield. Callers that move exactly
// one permanent use it; see Duration.Pinned.
//
// Caller must hold g.mu.
func (g *Game) PinnedTo(d Duration, object uuid.UUID) Duration {
	c, ok := g.battlefieldCardLocked(object)
	if !ok {
		return d
	}
	d.Pinned = object
	d.PinnedEnteredAt = c.EnteredBattlefieldAt
	return d
}

// durationExpiredLocked is the ONE function that decides whether a
// continuous effect is over. `endOfTurn` is true only in the CR 514.2
// cleanup sweep; every other caller passes false.
//
// Caller must hold g.mu.
func (g *Game) durationExpiredLocked(d Duration, endOfTurn bool) bool {
	// The pin comes first and applies to every kind: an effect that
	// exists to move one permanent is over when that permanent is
	// gone, or has come back as a different object (CR 400.7).
	if d.Pinned != uuid.Nil && !g.sameObjectOnBattlefieldLocked(d.Pinned, d.PinnedEnteredAt) {
		return true
	}
	switch d.Kind {
	case UntilEndOfTurn:
		// The cleanup step of the turn it was made in ends it
		// (CR 514.2) — and so does the beginning of any later turn,
		// which is the backstop for a turn that ended early without
		// a real cleanup step (ADR 0059 Decision 6).
		return endOfTurn || g.turnsBegunForLocked(d.Player) > d.ExpiresAfterTurnsBegun
	case UntilYourNextTurn:
		return g.turnsBegunForLocked(d.Player) >= d.ExpiresAtTurnsBegun
	case ForAsLongAs:
		return !g.durationConditionHoldsLocked(d)
	case Indefinite:
		return false
	}
	return false
}

// durationConditionHoldsLocked re-runs a ForAsLongAs condition against
// the board as the previous layer pass left it. Caller must hold g.mu.
func (g *Game) durationConditionHoldsLocked(d Duration) bool {
	c, ok := g.battlefieldCardLocked(d.Source)
	if !ok || c.EnteredBattlefieldAt != d.SourceEnteredAt {
		return false
	}
	if d.Condition == WhileYouControlSource && c.Controller != d.Player {
		return false
	}
	return true
}

// sameObjectOnBattlefieldLocked reports whether `id` is on the
// battlefield AND is still the object that entered at `enteredAt`
// (CR 400.7). A zero stamp means "don't care", which is what a card
// seeded by a fixture without the zone-move event carries.
//
// Caller must hold g.mu.
func (g *Game) sameObjectOnBattlefieldLocked(id uuid.UUID, enteredAt int64) bool {
	c, ok := g.battlefieldCardLocked(id)
	if !ok {
		return false
	}
	return enteredAt == 0 || c.EnteredBattlefieldAt == enteredAt
}

// turnsBegunForLocked is a player's seat-turn count, or 0 for an
// unknown player (a fixture's synthetic source, a seat that was never
// seated). Caller must hold g.mu.
func (g *Game) turnsBegunForLocked(player uuid.UUID) int {
	if p := g.playerByIDLocked(player); p != nil {
		return p.TurnsBegun
	}
	return 0
}

// String names the duration kind for census labels and test failures.
// An operator reading a refused restore point needs to be able to tell
// a Giant Growth from an Agent of Treachery.
func (k DurationKind) String() string {
	switch k {
	case UntilEndOfTurn:
		return "until end of turn"
	case UntilYourNextTurn:
		return "until your next turn"
	case ForAsLongAs:
		return "for as long as"
	case Indefinite:
		return "no stated duration"
	}
	return "unknown duration"
}
