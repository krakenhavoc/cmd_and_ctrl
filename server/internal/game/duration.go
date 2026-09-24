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
// S42 / #945 added the fifth kind and a second client: a granted
// cast permission (ADR 0066) carries a `Duration` too, swept through
// the same `durationExpiredLocked`, so the engine has ONE duration
// vocabulary rather than the permission's own three fields.
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

// DurationKind names the five shapes CR 611.2 gives a continuous
// effect — or, since #945, a granted cast permission — created by a
// resolved spell or ability.
type DurationKind int

const (
	// UntilEndOfTurn ends during the cleanup step of the turn the
	// effect was created in (CR 514.2) — including an effect created
	// during that turn's END step, which is not the end of the turn.
	// The zero value, so an unstamped duration behaves exactly as
	// the pre-S38 registry did.
	UntilEndOfTurn DurationKind = iota

	// UntilYourNextTurn ends as the named player's next turn begins
	// (CR 500.1: a turn begins with its untap step; CR 502.3 puts
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

	// WhileInZone is "for as long as this card remains in the zone it
	// is in" — airbend's "WHILE IT'S EXILED, its owner may cast it
	// for {2}", warp's "for as long as it remains exiled", and every
	// permission derived from a permanent on the battlefield.
	//
	// The fifth kind, added by #945 when granted cast permissions
	// (ADR 0066) moved onto this model. It is a real CR 611.2b
	// duration and not a synonym for Indefinite: the effect DOES
	// state when it ends, the statement is just about a zone rather
	// than about a turn.
	//
	// Nothing in the switch below ends it, and that is the whole
	// point. What ends it is CR 400.7 — the permission names
	// {instance, epoch}, so the moment the card leaves the zone it is
	// a new object and the grant no longer names it
	// (anyNamedObjectStillThereLocked sweeps the husk). A duration
	// that has to be re-checked against a zone every query would be a
	// second copy of that rule, and the two would drift.
	//
	// A ScopedStatic must not carry it: a continuous effect is about
	// objects on the battlefield and has no zone-bound husk to be
	// swept by. `durationExpiredLocked` therefore treats it exactly as
	// Indefinite, which is the safe half of that mistake — the layer
	// pass keeps such an effect rather than dropping it silently.
	WhileInZone

	// durationKindEnd is a sentinel, not a kind: every kind this binary
	// knows is below it. Keep it LAST. DurationKind is persisted as a
	// bare int (a restore point's `duration.Kind`), so a kind a newer
	// build adds is unknown to an older one, and restore refuses it
	// (ErrUnknownEffectKey, ADR 0041 P4) rather than restoring an
	// effect that never ends.
	durationKindEnd
)

// Known reports whether this binary can interpret k.
func (k DurationKind) Known() bool { return k >= 0 && k < durationKindEnd }

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

	// WhileYouControlSourceOnceItLands — "until that player loses
	// control of it" (CR 702.62a, suspend's haste), said of an
	// object that is still a SPELL ON THE STACK when the effect is
	// created.
	//
	// The third condition, added by #990, and the only one with a
	// GRACE PERIOD: it holds while the source sits on the stack,
	// because the permanent the effect is about does not exist yet
	// (CR 608.3 — a permanent spell becomes a permanent as it
	// resolves). Without that, the two conditions above are both
	// false for a spell mid-flight and the layer pass would sweep
	// the grant away a priority round before the creature it is for
	// ever arrives.
	//
	// It does NOT carry a battlefield-entry stamp, because there is
	// none to read at registration. CR 400.7 is answered by the
	// sweep instead, and answered strictly: a permanent that leaves
	// the battlefield is neither on the stack nor controlled by
	// anyone, so the condition is false at the next recompute and
	// the effect is dropped for good. A creature that dies and is
	// reanimated the same turn comes back without haste, which is
	// the right answer for the same reason a flickered Sower of
	// Temptation gives its creature back.
	//
	// Use it only for an effect created at ANNOUNCE about the
	// permanent the spell will become. An effect created while its
	// object is already on the battlefield wants
	// WhileYouControlSource, which is strictly tighter.
	WhileYouControlSourceOnceItLands

	// WhileSourceRemainsTapped — "for as long as ~ remains tapped"
	// (Rust Tick, Amber Prison; #1313, ADR 0058's 2026-09-23
	// amendment). WhileSourceOnBattlefield, and the source must still
	// be tapped. The layer listener bumps the layer version on
	// EventTapCard / EventUntapCard, so the recompute sweep sees the
	// source untap. Once the source has untapped the effect is over
	// for good, even if the source is tapped again later: CR 611.2b
	// says the effect "doesn't last forever", and the sweep is what
	// makes the end permanent.
	//
	// Appended, never inserted: the enum's integer values are written
	// into snapshot files.
	WhileSourceRemainsTapped

	// durationConditionEnd is a sentinel, not a condition. Keep it
	// LAST, for the reason durationKindEnd gives: an unknown condition
	// would otherwise fall through to a bare "source on battlefield".
	durationConditionEnd
)

// Known reports whether this binary can interpret c.
func (c DurationCondition) Known() bool { return c >= 0 && c < durationConditionEnd }

// Known reports whether this binary can interpret the duration: its
// kind and its condition. The condition is checked whatever the kind,
// because its zero value is a known condition and a non-zero one came
// from somewhere.
func (d Duration) Known() bool { return d.Kind.Known() && d.Condition.Known() }

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
	// Counting seat-turns rather than reading Turn.Round is the
	// whole point: Turn.Round counts ROUNDS, so all four seats in a
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

// UntilEndOfYourNextTurnDuration is "until the end of your next turn"
// (CR 611.2b) — Reckless Impulse, Wrenn's Resolve, Prosper's Mystic
// Arcanum. Caller must hold g.mu.
//
// NOT a fifth kind, and the reason is the point of the counter: "the
// end of your next turn" is the same boundary "until end of turn"
// names, one seat-turn later. So it is an UntilEndOfTurn stamped
// against a turn that has not happened yet — `TurnsBegun + 1` for the
// named player — and `durationExpiredLocked` reads the same case for
// both. A kind of its own would be a second copy of one rule.
//
// It is a strictly LONGER window than UntilYourNextTurn, and the two
// are easy to confuse: "until your next turn" ends as that turn
// BEGINS (CR 500.1), "until the end of your next turn" ends at that
// turn's cleanup step (CR 514.2). Reckless Impulse prints the second.
func (g *Game) UntilEndOfYourNextTurnDuration(player uuid.UUID) Duration {
	return Duration{
		Kind:                   UntilEndOfTurn,
		Player:                 player,
		ExpiresAfterTurnsBegun: g.turnsBegunForLocked(player) + 1,
	}
}

// WhileInZoneDuration is "for as long as this card remains in the
// zone" (CR 611.2b) — see the WhileInZone kind. Needs no game state,
// because the zone half is enforced by the CR 400.7 object identity
// the permission already names, so it is a plain function.
func WhileInZoneDuration() Duration { return Duration{Kind: WhileInZone} }

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

// ForAsLongAsSourceTappedDuration is "for as long as ~ remains tapped"
// (CR 611.2b). Returns false when the source is not on the battlefield
// or is not tapped: the duration never starts, so the caller records
// nothing. Caller must hold g.mu.
func (g *Game) ForAsLongAsSourceTappedDuration(source uuid.UUID) (Duration, bool) {
	c, ok := g.battlefieldCardLocked(source)
	if !ok || !c.Tapped {
		return Duration{}, false
	}
	return Duration{
		Kind:            ForAsLongAs,
		Condition:       WhileSourceRemainsTapped,
		Source:          source,
		SourceEnteredAt: c.EnteredBattlefieldAt,
	}, true
}

// UntilYouLoseControlOfDuration is "until that player loses control
// of it" (CR 611.2b, and CR 702.62a's haste), for an effect created
// while `source` is still a spell on the stack. See
// WhileYouControlSourceOnceItLands for the grace period that makes
// that legal and for the CR 400.7 reading of a permanent that leaves.
//
// Needs no game state — there is no entry stamp to read yet — so it
// is a plain function rather than a method.
func UntilYouLoseControlOfDuration(source, player uuid.UUID) Duration {
	return Duration{
		Kind:      ForAsLongAs,
		Condition: WhileYouControlSourceOnceItLands,
		Player:    player,
		Source:    source,
	}
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
	//
	// A PHASED-OUT permanent is neither (#1497 review): phasing is not
	// a zone change (CR 702.26d), the object keeps its instance and
	// its entry stamp, and an indefinite effect on it — an earthbend,
	// Tree of Perdition's toughness, Kyoshi's Island, a theft — applies
	// again when it phases in. While it is out the layer pass cannot
	// see it, so the effect affects nothing in the meantime, which is
	// CR 702.26e's exclusion from the affected set. The pin is
	// garbage collection, not a "for as long as" duration, so it is
	// the one check that looks in g.PhasedOut: a ForAsLongAs condition
	// below still reads the battlefield alone, because CR 702.26f ends
	// a duration that tracks a permanent once it phases out.
	if d.Pinned != uuid.Nil && !g.sameObjectPresentLocked(d.Pinned, d.PinnedEnteredAt) {
		return true
	}
	switch d.Kind {
	case UntilEndOfTurn:
		// The cleanup step of the turn the effect NAMES ends it
		// (CR 514.2) — and so does the beginning of any turn after
		// that one, which is the backstop for a turn that ended early
		// without a real cleanup step (ADR 0059 Decision 6).
		//
		// The turn it names is the one whose ExpiresAfterTurnsBegun
		// the stamp carries: the creating player's CURRENT turn for
		// UntilEndOfTurnDuration, their NEXT turn for
		// UntilEndOfYourNextTurnDuration (#945). For every effect
		// stamped by the first constructor the seat-turn test is
		// already true when the sweep runs, so this reads exactly as
		// the pre-#945 `endOfTurn ||` did.
		turns := g.turnsBegunForLocked(d.Player)
		return (endOfTurn && turns >= d.ExpiresAfterTurnsBegun) || turns > d.ExpiresAfterTurnsBegun
	case UntilYourNextTurn:
		return g.turnsBegunForLocked(d.Player) >= d.ExpiresAtTurnsBegun
	case ForAsLongAs:
		return !g.durationConditionHoldsLocked(d)
	case Indefinite, WhileInZone:
		// Neither ends on a turn boundary. WhileInZone ends on a ZONE
		// change, which CR 400.7 answers on the permission itself —
		// see the kind's comment.
		return false
	}
	return false
}

// durationConditionHoldsLocked re-runs a ForAsLongAs condition against
// the board as the previous layer pass left it. Caller must hold g.mu.
func (g *Game) durationConditionHoldsLocked(d Duration) bool {
	if d.Condition == WhileYouControlSourceOnceItLands {
		// #990: the object may not be a permanent yet. On the stack
		// the condition holds (the effect has not started applying to
		// anything), on the battlefield it is the ordinary control
		// test, and anywhere else it is over. No entry-stamp
		// comparison, because the grant was made before there was one
		// — see the condition's own comment for why that is the
		// STRICTER reading of CR 400.7 rather than the looser one.
		if c, ok := g.battlefieldCardLocked(d.Source); ok {
			return c.Controller == d.Player
		}
		return g.Stack != nil && g.Stack.Contains(d.Source)
	}
	c, ok := g.battlefieldCardLocked(d.Source)
	if !ok || c.EnteredBattlefieldAt != d.SourceEnteredAt {
		return false
	}
	if d.Condition == WhileYouControlSource && c.Controller != d.Player {
		return false
	}
	if d.Condition == WhileSourceRemainsTapped && !c.Tapped {
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

// sameObjectPresentLocked is sameObjectOnBattlefieldLocked that also
// accepts the same object phased out (CR 702.26d: phasing is not a
// zone change, so a phased-out permanent is still that object). Used
// only by the pin — see durationExpiredLocked for why the ForAsLongAs
// conditions deliberately do not use it.
//
// Caller must hold g.mu.
func (g *Game) sameObjectPresentLocked(id uuid.UUID, enteredAt int64) bool {
	if g.sameObjectOnBattlefieldLocked(id, enteredAt) {
		return true
	}
	if g.PhasedOut == nil {
		return false
	}
	for _, c := range g.PhasedOut.Cards {
		if c.InstanceID == id {
			return enteredAt == 0 || c.EnteredBattlefieldAt == enteredAt
		}
	}
	return false
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
	case WhileInZone:
		return "while it remains in the zone"
	}
	return "unknown duration"
}
