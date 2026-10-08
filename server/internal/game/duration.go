package game

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

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
//   - `Clone` copies a `ScopedEffect` by value, so a duration that
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
	// A ScopedEffect must not carry it: a continuous effect is about
	// objects on the battlefield and has no zone-bound husk to be
	// swept by. `durationExpiredLocked` therefore treats it exactly as
	// Indefinite, which is the safe half of that mistake — the layer
	// pass keeps such an effect rather than dropping it silently.
	WhileInZone

	// UntilYourNextEndStep ends as the named player's NEXT end step
	// begins (#2373): "exile the top card, until your next end step you
	// may play it" — Ob Nixilis, Captive Kingpin, Haste Magic, Wiccan.
	//
	// The sixth kind, appended after WhileInZone so every persisted
	// int keeps its meaning. A restore point written before it never
	// carries it; one written after it is refused by an older binary
	// through Known, never restored as a window that does not end.
	//
	// Which end step is "next" is stamped at creation
	// (UntilYourNextEndStepDuration): this turn's when it is the
	// player's own turn and their end step has not begun, otherwise
	// the end step of their next turn. The boundary is the BEGINNING
	// of that step, so a card exiled for it cannot be played in the
	// end step itself (CR 611.2b: the effect ends as the step begins).
	UntilYourNextEndStep

	// UntilEndOfCombat is "this combat" and "until end of combat"
	// (ADR 0108 amendment 2026-10-07, #2027): the effect lasts until the
	// combat PHASE it was made in ends. CR 511.3 and 724.2d both say
	// "effects that last 'until end of combat' expire", at the moment
	// the end of combat step ends or an effect ends the phase early.
	//
	// The seventh kind, appended after UntilYourNextEndStep so every
	// persisted int keeps its meaning. A binary from before it refuses
	// a file carrying it through Known; it never restores a window that
	// does not end.
	//
	// Which combat is "this combat" is the phase INSTANCE, not the turn:
	// an additional combat phase (Aurelia, Karlach) is a second combat
	// in the same turn, and a "this combat" effect from the first has
	// ended by the time the second begins. The stamp is the turn's
	// identity (Turn.Seq) plus the phase instance (Turn.PhaseID), read
	// by UntilEndOfCombatDuration and compared with the cursor by
	// durationExpiredLocked. It is a pure read of the cursor, so there
	// is no flag to forget to clear.
	UntilEndOfCombat

	// UntilSourceExilesAnother is "you may play that card until you
	// exile another card with this artifact" (ADR 0066 amendment
	// 2026-10-08, #2539): Unstable Amulet, Furious Rise, Superior Foes
	// of Spider-Man. The window ends on an EVENT, not a turn boundary
	// (CR 611.2a): the holder's next exile made by the same linked
	// ability of the same source OBJECT (CR 607.2a, CR 607.1c).
	//
	// The eighth kind, appended after UntilEndOfCombat so every
	// persisted int keeps its meaning. A binary from before it refuses
	// a file carrying it through Known, so a rollback never restores the
	// window as one that does not end.
	//
	// The duration names its holder (Player) and the source object
	// (Source plus SourceEpoch, the shape of StackItem.SourceObject).
	// ExileTopUntilAnotherForEffect, the one helper that exiles "with
	// this", sets Ended on every earlier window it closes, and
	// durationExpiredLocked reads only that flag. Nothing else ends it:
	// a source that leaves the battlefield leaves the newest card
	// playable for as long as it stays exiled (the Unstable Amulet and
	// Furious Rise rulings), and the card leaving exile ends the
	// permission by object identity (ADR 0066 decision 2).
	UntilSourceExilesAnother

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

	// WhilePinnedHasCounter — "for as long as that <object> has a
	// <kind> counter on it" (ADR 0109 §2, #1604): Aquitect's Will's
	// flood counter, Shield Broker's shield counter. It reads the
	// PINNED object, not the source: the pinned permanent must still be
	// on the battlefield as the same object (CR 400.7) with at least one
	// counter of Duration.CounterKind. The layer listener bumps the
	// layer version on EventCounterPlaced, which also fires on removal,
	// so the next sweep sees the last counter go. Once it has gone the
	// effect is over for good (CR 611.2b), even if a counter of that
	// kind is put on the permanent again later.
	WhilePinnedHasCounter

	// WhilePinnedRemainsTapped — "for as long as that creature remains
	// tapped" (ADR 0109 §3, #1894): Zygon Infiltrator's copy lasts while
	// the creature it tapped stays tapped. WhileSourceRemainsTapped
	// about the pinned object instead of the source.
	WhilePinnedRemainsTapped

	// WhilePinnedPowerAtMostSource — "and that creature's power remains
	// less than or equal to this creature's power" (ADR 0109 §3, Old Man
	// of the Sea). Both powers are read live, from the layer pass, and
	// the source must still be the object the duration names. The
	// recompute re-checks it after every pass, because a power change is
	// a layer OUTPUT that nothing else re-sweeps (powerConditionsFailLocked).
	WhilePinnedPowerAtMostSource

	// durationConditionEnd is a sentinel, not a condition. Keep it
	// LAST, for the reason durationKindEnd gives: an unknown condition
	// would otherwise fall through to a bare "source on battlefield".
	durationConditionEnd
)

// Known reports whether this binary can interpret c.
func (c DurationCondition) Known() bool { return c >= 0 && c < durationConditionEnd }

// readsPin reports whether c is about the pinned object, so a duration
// using it must name one.
func (c DurationCondition) readsPin() bool {
	return c == WhilePinnedHasCounter || c == WhilePinnedRemainsTapped || c == WhilePinnedPowerAtMostSource
}

// Known reports whether this binary can interpret the duration: its
// kind, its condition and every condition joined to it, and the fields
// those conditions read. The condition is checked whatever the kind,
// because its zero value is a known condition and a non-zero one came
// from somewhere.
//
// It is the restore check for every stored duration (ADR 0109 Shared
// machinery 2), so it fails CLOSED: a duration this binary would read
// differently from the binary that wrote it is unknown, not restored
// with a field ignored. See Problem.
func (d Duration) Known() bool { return d.Problem() == "" }

// Problem says why this binary cannot interpret d, or "" when it can.
//
//   - An unknown kind or condition is a newer binary's vocabulary.
//   - Also (ADR 0109 §3) is read only by ForAsLongAs; on any other kind
//     this binary would ignore it, which is the effect outliving a
//     condition the writer meant it to have.
//   - CounterKind is read only by WhilePinnedHasCounter, and that
//     condition is meaningless without one.
//   - A condition about the pinned object needs a pin.
func (d Duration) Problem() string {
	if !d.Kind.Known() {
		return fmt.Sprintf("duration kind %d", d.Kind)
	}
	if !d.Condition.Known() {
		return fmt.Sprintf("duration condition %d", d.Condition)
	}
	for _, c := range d.Also {
		if !c.Known() {
			return fmt.Sprintf("duration condition %d joined by Also", c)
		}
	}
	if d.Kind != UntilEndOfCombat && (d.CombatTurn != 0 || d.CombatPhase != 0) {
		return fmt.Sprintf("a combat stamp on a %s duration", d.Kind)
	}
	if d.Kind != UntilSourceExilesAnother && (d.SourceEpoch != 0 || d.Ended) {
		return fmt.Sprintf("a source-exile stamp on a %s duration", d.Kind)
	}
	if d.Kind == UntilSourceExilesAnother && d.Source == uuid.Nil {
		return "an until-you-exile-another duration names no source"
	}
	if d.Kind != ForAsLongAs {
		if len(d.Also) > 0 {
			return fmt.Sprintf("Also on a %s duration", d.Kind)
		}
		if d.CounterKind != "" {
			return fmt.Sprintf("CounterKind %q on a %s duration", d.CounterKind, d.Kind)
		}
		return ""
	}
	counter, pin := false, false
	for _, c := range d.conditions() {
		counter = counter || c == WhilePinnedHasCounter
		pin = pin || c.readsPin()
	}
	switch {
	case counter && d.CounterKind == "":
		return "a counter-held duration names no counter kind"
	case !counter && d.CounterKind != "":
		return fmt.Sprintf("CounterKind %q with no counter-held condition", d.CounterKind)
	case pin && d.Pinned == uuid.Nil:
		return "a duration about the pinned object names no pin"
	}
	return ""
}

// conditions is Condition followed by every condition in Also: the
// conjunction a ForAsLongAs duration holds while (CR 611.2b).
func (d Duration) conditions() []DurationCondition {
	return append([]DurationCondition{d.Condition}, d.Also...)
}

// And returns a copy of d that also lasts only while each of `conds`
// holds (ADR 0109 §3): "for as long as you control this creature AND
// this creature remains tapped". Only meaningful on a ForAsLongAs
// duration; Problem refuses it on any other kind. The receiver's Also
// is never written through, because a Duration is shared by value with
// every undo snapshot.
func (d Duration) And(conds ...DurationCondition) Duration {
	if len(conds) == 0 {
		return d
	}
	d.Also = append(append([]DurationCondition(nil), d.Also...), conds...)
	return d
}

// IsZero reports whether d is the zero Duration — "until end of turn"
// with no stamp, which several callers read as "no duration given".
func (d Duration) IsZero() bool { return d.Equal(Duration{}) }

// Equal reports whether two durations are the same duration. A Duration
// is not comparable with == since Also is a slice.
// An empty Also and a nil one are the same conjunction.
func (d Duration) Equal(o Duration) bool {
	if len(d.Also) == 0 {
		d.Also = nil
	}
	if len(o.Also) == 0 {
		o.Also = nil
	}
	return reflect.DeepEqual(d, o)
}

// Duration is how long one continuous effect lasts. The zero value is
// "until end of turn" with no stamp, which the sweep treats as ending
// at the first cleanup it sees — the pre-S38 behaviour.
//
// IMMUTABLE after registration, like every other field on
// ScopedEffect: `Clone` shares the value with every undo snapshot.
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

	// PinnedUnstamped is AffectedObject.Unstamped for the pin (#1558):
	// the pinned permanent had no entry stamp, so the pin names it
	// exactly — "still unstamped" — rather than falling back to the
	// PinnedEnteredAt 0 wildcard, which went on matching the object's
	// successor after a flicker. Omitted when false, so every duration
	// written before it reads the same; an older binary that ignores
	// it gets the pre-#1558 wildcard back.
	PinnedUnstamped bool `json:"PinnedUnstamped,omitempty"`

	// PinnedOnStack and PinnedEpoch are the pin for a SPELL (ADR 0104,
	// #1745): Pinned names an object on the stack, as the instance it
	// is now plus the Card.ObjectEpoch it has there. The effect ends
	// once that object is no longer on the stack, which is what
	// collects the record of a stolen spell that is countered or
	// resolves as an instant. A stolen PERMANENT spell's record is
	// re-pinned to the permanent as it lands (CR 400.7a), before this
	// check can see it gone. Omitted when false, so every duration
	// written before them reads the same.
	//
	// PinnedEpoch without PinnedOnStack (and with PinnedEnteredAt 0 and
	// PinnedUnstamped false) pins a PERMANENT by its ObjectEpoch rather
	// than its entry stamp — PinnedToEpoch, AffectedObject's
	// PinObjectByEpoch on the duration side.
	PinnedOnStack bool `json:"PinnedOnStack,omitempty"`
	PinnedEpoch   int  `json:"PinnedEpoch,omitempty"`

	// CounterKind is the counter a WhilePinnedHasCounter duration
	// watches ("flood", "shield"; ADR 0109 §2). Empty for every other
	// duration, and omitted then, so every duration written before it
	// reads the same. A binary before it refuses a record carrying it:
	// the unknown-field scan names it, and Known refuses the condition.
	CounterKind string `json:"CounterKind,omitempty"`

	// CombatTurn and CombatPhase name the combat an UntilEndOfCombat
	// duration lasts through: Turn.Seq and Turn.PhaseID as it was made
	// (see the kind). Zero, and omitted, for every other kind; Problem
	// refuses them there, because a binary that ignored them would be
	// reading a duration differently from the one that wrote it.
	CombatTurn  int `json:"CombatTurn,omitempty"`
	CombatPhase int `json:"CombatPhase,omitempty"`

	// SourceEpoch is the Card.ObjectEpoch of the object an
	// UntilSourceExilesAnother duration names, beside its instance in
	// Source: together they are StackItem.SourceObject, so CR 400.7 is
	// checked rather than assumed (an Amulet that left and came back
	// is a new object, and its exiles close none of the old one's
	// windows). Ended is set once that object's linked ability has
	// exiled another card for the same holder (ADR 0066 amendment
	// 2026-10-08). Both are zero, and omitted, for every other kind;
	// Problem refuses them there.
	SourceEpoch int  `json:"SourceEpoch,omitempty"`
	Ended       bool `json:"Ended,omitempty"`

	// Also is the rest of a conjunction (ADR 0109 §3, CR 611.2b): a
	// ForAsLongAs duration lasts while Condition AND every condition
	// listed here hold, and ends the moment any one of them stops.
	// Seasinger's "for as long as you control this creature and this
	// creature remains tapped" is WhileYouControlSource with
	// WhileSourceRemainsTapped here. A list rather than a combined
	// condition per pair, because each card prints a different pair.
	// Omitted when empty. IMMUTABLE once registered, like the rest of
	// the value: build it with And, which never writes through.
	Also []DurationCondition `json:"Also,omitempty"`
}

// PinnedToEpoch is PinnedTo for a permanent named by its ObjectEpoch
// (ADR 0104): the pin a stolen permanent spell's record takes as the
// permanent lands, before the entry stamp exists. `d` is returned
// unchanged when `object` is not on the battlefield or has no epoch.
//
// Caller must hold g.mu.
func (g *Game) PinnedToEpoch(d Duration, object uuid.UUID) Duration {
	c, ok := g.battlefieldCardLocked(object)
	if !ok || c.ObjectEpoch <= 0 {
		return d
	}
	d.Pinned = object
	d.PinnedEnteredAt = 0
	d.PinnedUnstamped = false
	d.PinnedOnStack = false
	d.PinnedEpoch = c.ObjectEpoch
	return d
}

// PinnedToStack is PinnedTo for a spell on the stack (ADR 0104): the
// duration also ends when that object leaves the stack. `d` is
// returned unchanged when `object` is not a card on the stack.
//
// Caller must hold g.mu.
func (g *Game) PinnedToStack(d Duration, object uuid.UUID) Duration {
	c, ok := g.stackCardLocked(object)
	if !ok {
		return d
	}
	d.Pinned = object
	d.PinnedEnteredAt = 0
	d.PinnedUnstamped = false
	d.PinnedOnStack = true
	d.PinnedEpoch = c.ObjectEpoch
	return d
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

// UntilYourNextEndStepDuration is "until your next end step"
// (CR 611.2b, #2373), stamped against `player`'s seat-turn count.
// Caller must hold g.mu.
//
// From the player's own turn before its end step (any step except end
// and cleanup) the next end step is this turn's, so the stamp is their
// CURRENT turn. From their end step or cleanup — that end step has
// already begun — or from anyone else's turn, it is their next turn's,
// so the stamp is `TurnsBegun + 1`. `ExpiresAtTurnsBegun` is the turn
// whose end step ends the effect; durationExpiredLocked reads it.
func (g *Game) UntilYourNextEndStepDuration(player uuid.UUID) Duration {
	target := g.turnsBegunForLocked(player)
	if !g.endStepOfTurnStillAheadLocked(player) {
		target++
	}
	return Duration{
		Kind:                UntilYourNextEndStep,
		Player:              player,
		ExpiresAtTurnsBegun: target,
	}
}

// UntilEndOfCombatDuration is "this combat" / "until end of combat"
// (CR 511.3, 724.2d), stamped against the combat phase in progress.
// The second return is false outside a combat phase: an effect that
// lasts "this combat" when there is no combat has no combat to last
// through, so by CR 611.2b's reading it never begins and the caller
// registers nothing. Caller must hold g.mu.
func (g *Game) UntilEndOfCombatDuration() (Duration, bool) {
	if PhaseOf(g.Turn.Step) != PhaseCombat {
		return Duration{}, false
	}
	return Duration{
		Kind:        UntilEndOfCombat,
		CombatTurn:  g.Turn.Seq,
		CombatPhase: g.Turn.PhaseID,
	}, true
}

// inCombatPhaseLocked reports whether the cursor is inside the combat
// phase instance named by (turn, phase). Caller must hold g.mu.
func (g *Game) inCombatPhaseLocked(turn, phase int) bool {
	return PhaseOf(g.Turn.Step) == PhaseCombat && g.Turn.Seq == turn && g.Turn.PhaseID == phase
}

// endStepOfTurnStillAheadLocked reports whether it is `player`'s own
// turn and their end step has not begun. Caller must hold g.mu.
func (g *Game) endStepOfTurnStillAheadLocked(player uuid.UUID) bool {
	if player == uuid.Nil || g.activePlayerIDLocked() != player {
		return false
	}
	return !g.stepAtOrAfterEndLocked()
}

// stepAtOrAfterEndLocked is true during the end and cleanup steps.
// Caller must hold g.mu.
func (g *Game) stepAtOrAfterEndLocked() bool {
	return g.Turn.Step == StepEnd || g.Turn.Step == StepCleanup
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

// ForAsLongAsYouControlAndSourceTappedDuration is "for as long as you
// control ~ and ~ remains tapped" (ADR 0109 §3, #1894): Seasinger,
// Rubinia Soulsinger, Willow Satyr, Hivis of the Scale, Helm of
// Possession. WhileYouControlSource joined with WhileSourceRemainsTapped;
// the effect ends the moment either stops (CR 611.2b). Returns false,
// so the caller registers nothing, when either half is already false.
//
// Caller must hold g.mu.
func (g *Game) ForAsLongAsYouControlAndSourceTappedDuration(source, player uuid.UUID) (Duration, bool) {
	d, ok := g.ForAsLongAsYouControlDuration(source, player)
	if !ok {
		return Duration{}, false
	}
	if c, _ := g.battlefieldCardLocked(source); !c.Tapped {
		return Duration{}, false
	}
	return d.And(WhileSourceRemainsTapped), true
}

// ForAsLongAsPinnedHasCounterDuration is "for as long as that <object>
// has a <kind> counter on it" (ADR 0109 §2, #1604), pinned to `object`
// as the object it is now (CR 400.7). Returns false when `object` is
// not on the battlefield or has no such counter: the duration never
// starts (CR 611.2b), so the caller registers nothing. A card puts the
// counter first and builds the duration second, in the order it prints.
//
// Caller must hold g.mu.
func (g *Game) ForAsLongAsPinnedHasCounterDuration(object uuid.UUID, kind string) (Duration, bool) {
	c, ok := g.battlefieldCardLocked(object)
	if !ok || kind == "" || c.Counters[kind] <= 0 {
		return Duration{}, false
	}
	return g.PinnedTo(Duration{Kind: ForAsLongAs, Condition: WhilePinnedHasCounter, CounterKind: kind}, object), true
}

// ForAsLongAsPinnedTappedDuration is "for as long as that creature
// remains tapped" (ADR 0109 §3): Zygon Infiltrator, about the creature
// it tapped. Pinned to `object`; false, so nothing is registered, when
// `object` is not on the battlefield or is untapped.
//
// Caller must hold g.mu.
func (g *Game) ForAsLongAsPinnedTappedDuration(object uuid.UUID) (Duration, bool) {
	c, ok := g.battlefieldCardLocked(object)
	if !ok || !c.Tapped {
		return Duration{}, false
	}
	return g.PinnedTo(Duration{Kind: ForAsLongAs, Condition: WhilePinnedRemainsTapped}, object), true
}

// ForAsLongAsSourceTappedAndPowerAtMostDuration is Old Man of the Sea's
// "for as long as this creature remains tapped and that creature's
// power remains less than or equal to this creature's power" (ADR 0109
// §3): WhileSourceRemainsTapped joined with WhilePinnedPowerAtMostSource,
// pinned to `object`. False when either half is already false.
//
// Caller must hold g.mu.
func (g *Game) ForAsLongAsSourceTappedAndPowerAtMostDuration(source, object uuid.UUID) (Duration, bool) {
	d, ok := g.ForAsLongAsSourceTappedDuration(source)
	if !ok {
		return Duration{}, false
	}
	if _, ok := g.battlefieldCardLocked(object); !ok {
		return Duration{}, false
	}
	d = g.PinnedTo(d.And(WhilePinnedPowerAtMostSource), object)
	if !g.durationConditionHoldsLocked(d) {
		return Duration{}, false
	}
	return d, true
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
	d.PinnedUnstamped = c.EnteredBattlefieldAt == 0
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
	if d.Pinned != uuid.Nil && d.PinnedOnStack {
		// ADR 0104: a spell's pin. Over once that object has left the
		// stack, whatever the kind says.
		if c, ok := g.stackCardLocked(d.Pinned); !ok || c.ObjectEpoch != d.PinnedEpoch {
			return true
		}
	} else if d.Pinned != uuid.Nil && d.PinnedEpoch > 0 && d.PinnedEnteredAt == 0 && !d.PinnedUnstamped {
		// ADR 0104: a permanent pinned by its epoch. Phasing keeps the
		// object (CR 702.26d), so a phased-out one is still it.
		if !g.sameEpochPresentLocked(d.Pinned, d.PinnedEpoch) {
			return true
		}
	} else if d.Pinned != uuid.Nil && !g.sameObjectPresentLocked(d.Pinned, d.PinnedEnteredAt, d.PinnedUnstamped) {
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
	case UntilYourNextEndStep:
		// ExpiresAtTurnsBegun names the player's turn whose end step
		// closes the window, and the window closes as an end step
		// begins — whatever the sweep has or has not seen, because a
		// cast permission is read live.
		//
		// A turn being OVER is not "its end step began" (#2385): a turn that is
		// ended (CR 724.1) skips its end step, which then never
		// begins, so the window carries to the player's next end step
		// that does. Player.EndStepTurn is the seat-turn count of the
		// last end step that began, so any real end step on or after
		// the stamped turn closes it, and a turn that skipped it does
		// not.
		p := g.playerByIDLocked(d.Player)
		return p != nil && p.EndStepTurn >= d.ExpiresAtTurnsBegun
	case UntilEndOfCombat:
		// Over the moment the cursor is outside the combat phase it
		// names (CR 511.3, 724.2d): the next step after end of combat,
		// the next phase an effect ended the combat for, or any later
		// turn. Pure, so the sweeps that run at step transitions
		// (advanceCursorLocked) and at cleanup need no special case.
		return !g.inCombatPhaseLocked(d.CombatTurn, d.CombatPhase)
	case ForAsLongAs:
		return !g.durationConditionHoldsLocked(d)
	case UntilSourceExilesAnother:
		// Over once the linked exile marked it (#2539). The source
		// leaving the battlefield does not end it, which is the
		// ruling; the card leaving exile ends the permission through
		// its object check, not here.
		return d.Ended
	case Indefinite, WhileInZone:
		// Neither ends on a turn boundary. WhileInZone ends on a ZONE
		// change, which CR 400.7 answers on the permission itself —
		// see the kind's comment.
		return false
	}
	return false
}

// durationConditionHoldsLocked re-runs a ForAsLongAs duration's
// conditions against the board as the previous layer pass left it:
// Condition and every condition in Also, all of which must hold
// (CR 611.2b; ADR 0109 §3). Caller must hold g.mu.
func (g *Game) durationConditionHoldsLocked(d Duration) bool {
	for _, c := range d.conditions() {
		if !g.conditionHoldsLocked(d, c) {
			return false
		}
	}
	return true
}

// conditionHoldsLocked is one condition of d's conjunction. An unknown
// condition never holds: restore refuses one before it can get here,
// and failing closed ends the effect rather than keeping it forever.
//
// Caller must hold g.mu.
func (g *Game) conditionHoldsLocked(d Duration, cond DurationCondition) bool {
	switch cond {
	case WhilePinnedHasCounter:
		c, ok := g.pinnedOnBattlefieldLocked(d)
		return ok && d.CounterKind != "" && c.Counters[d.CounterKind] > 0
	case WhilePinnedRemainsTapped:
		c, ok := g.pinnedOnBattlefieldLocked(d)
		return ok && c.Tapped
	case WhilePinnedPowerAtMostSource:
		c, ok := g.pinnedOnBattlefieldLocked(d)
		if !ok {
			return false
		}
		src, ok := g.battlefieldCardLocked(d.Source)
		if !ok || src.EnteredBattlefieldAt != d.SourceEnteredAt {
			return false
		}
		return c.CurrentPower() <= src.CurrentPower()
	case WhileSourceOnBattlefield, WhileYouControlSource, WhileYouControlSourceOnceItLands, WhileSourceRemainsTapped:
		return g.sourceConditionHoldsLocked(d, cond)
	}
	return false
}

// pinnedOnBattlefieldLocked is the pinned object of d, if it is on the
// battlefield as the object the pin names (CR 400.7). Unlike the pin's
// own garbage collection it does not look among the phased-out: a
// condition about a permanent ends when it phases out (CR 702.26f).
//
// Caller must hold g.mu.
func (g *Game) pinnedOnBattlefieldLocked(d Duration) (*Card, bool) {
	if d.Pinned == uuid.Nil || d.PinnedOnStack {
		return nil, false
	}
	c, ok := g.battlefieldCardLocked(d.Pinned)
	if !ok {
		return nil, false
	}
	if d.PinnedEpoch > 0 && d.PinnedEnteredAt == 0 && !d.PinnedUnstamped {
		return c, c.ObjectEpoch == d.PinnedEpoch
	}
	return c, entryMatches(d.PinnedEnteredAt, d.PinnedUnstamped, c.EnteredBattlefieldAt)
}

// powerConditionsFailLocked reports whether a scoped effect whose
// duration compares powers (WhilePinnedPowerAtMostSource) no longer
// holds. Power is a layer OUTPUT: the sweep at the top of a recompute
// reads the previous pass, so a pass that lowers the source's power, or
// raises the pinned creature's, would otherwise leave the effect in
// place until something unrelated bumped the layer version. The
// recompute asks this after its pass and bumps the version when it
// says true, so the next read ends the effect.
//
// Caller must hold g.mu.
func (g *Game) powerConditionsFailLocked() bool {
	for _, e := range g.ScopedEffects {
		if e.Duration.Kind != ForAsLongAs {
			continue
		}
		for _, c := range e.Duration.conditions() {
			if c == WhilePinnedPowerAtMostSource && !g.conditionHoldsLocked(e.Duration, c) {
				return true
			}
		}
	}
	return false
}

// DurationHoldsForEffect reports whether d is still running: the
// question a card asks of a duration it has just built from parts,
// because CR 611.2b says an effect whose duration has already ended as
// it would begin never begins. Caller must hold g.mu.
func (g *Game) DurationHoldsForEffect(d Duration) bool {
	return !g.durationExpiredLocked(d, false)
}

// sourceConditionHoldsLocked is the four conditions about the source.
// Caller must hold g.mu.
func (g *Game) sourceConditionHoldsLocked(d Duration, cond DurationCondition) bool {
	if cond == WhileYouControlSourceOnceItLands {
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
	if cond == WhileYouControlSource && c.Controller != d.Player {
		return false
	}
	if cond == WhileSourceRemainsTapped && !c.Tapped {
		return false
	}
	return true
}

// sameObjectOnBattlefieldLocked reports whether `id` is on the
// battlefield AND is still the object that entered at `enteredAt`
// (CR 400.7), read by entryMatches: `enteredAt` 0 means "don't care",
// and `unstamped` means "still the unstamped object a fixture seeded
// without the zone-move event" (#1558).
//
// Caller must hold g.mu.
func (g *Game) sameObjectOnBattlefieldLocked(id uuid.UUID, enteredAt int64, unstamped bool) bool {
	c, ok := g.battlefieldCardLocked(id)
	if !ok {
		return false
	}
	return entryMatches(enteredAt, unstamped, c.EnteredBattlefieldAt)
}

// sameObjectPresentLocked is sameObjectOnBattlefieldLocked that also
// accepts the same object phased out (CR 702.26d: phasing is not a
// zone change, so a phased-out permanent is still that object). Used
// only by the pin — see durationExpiredLocked for why the ForAsLongAs
// conditions deliberately do not use it.
//
// Caller must hold g.mu.
func (g *Game) sameObjectPresentLocked(id uuid.UUID, enteredAt int64, unstamped bool) bool {
	if g.sameObjectOnBattlefieldLocked(id, enteredAt, unstamped) {
		return true
	}
	if g.PhasedOut == nil {
		return false
	}
	for _, c := range g.PhasedOut.Cards {
		if c.InstanceID == id {
			return entryMatches(enteredAt, unstamped, c.EnteredBattlefieldAt)
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
	case UntilYourNextEndStep:
		return "until your next end step"
	case UntilEndOfCombat:
		return "until end of combat"
	case UntilSourceExilesAnother:
		return "until you exile another card with it"
	}
	return "unknown duration"
}
