package game

import (
	"errors"

	"github.com/google/uuid"
)

// earthbend.go — the CR 701 keyword action Earthbend N (#1178).
//
//	Earthbend N. (Target land you control becomes a 0/0 creature with
//	haste that's still a land. Put N +1/+1 counters on it. When it
//	dies or is exiled, return it to the battlefield tapped.)
//
// Twelve cards in the Avatar set print it (Bitter Work, Rockalanche,
// Earth Rumble, Bumi King of Three Trials, Ba Sing Se, Sandbenders'
// Storm, Cracked Earth Technique, Earthshape, Dai Li Indoctrination,
// Earthbending Lesson, The Cave of Two Lovers and The Legend of
// Kyoshi's chapter II), and every one of them prints the same four
// sentences. So the four are ONE verb here and each card is one
// clause, in the shape ADR 0013 §5s gave proliferate, scry and
// surveil: a `KeywordAction` constant, a count on the CR 614 event,
// and one arm of `applyResolvedKeywordActionLocked`.
//
// # Why it is a keyword action and not a card primitive
//
// The count is printed. "Earthbend 4", "earthbend X where X is the
// number of Forests you control", "earthbend 2, then earthbend 2" —
// the number is the only thing that varies across the twelve, which
// is exactly the shape a CR 614 window is for. Opening one costs
// nothing today (nothing printed replaces an earthbend) and is the
// difference between a seam and a rewrite the day one does: a
// "whenever you would earthbend, earthbend twice that much instead"
// is a card-side `ReplacementEffect` narrowing on the action, with no
// engine change at all.
//
// Earthbend is the FIRST counted keyword action whose count is a
// number of COUNTERS rather than a number of times (proliferate) or a
// number of cards (scry, surveil) — see `KeywordActionEarthbend`. It
// is also the first that DOES SOMETHING AT ZERO, which is the one
// place the shared body had to learn a new fact; see
// `KeywordAction.actsAtZeroCount`.
//
// # The four parts, and where each one already lived
//
//  1. The ANIMATION is two continuous effects with one CR 613.7
//     timestamp: layer 4 adds Creature (Land is kept — the card says
//     "that's still a land", and adding a type never removes one),
//     layer 7b sets base power and toughness to 0/0. `BecomeCreature-
//     UntilEOT` is the card-side shape for the same idea and could
//     not be reused: its zero P/T means "leave the printed values
//     alone", and its own comment says a card that really wants a 0/0
//     "is not expressible and does not exist". This is that card.
//  2. HASTE is a third continuous effect at layer 6, sharing the
//     animation's duration and timestamp, because it is the same
//     printed sentence. NOT an until-end-of-turn grant: a land
//     earthbent on turn three is still hasty on turn nine.
//  3. The COUNTERS go through `AddCounterByForEffect`, so the CR 614
//     placement window opens and Hardened Scales, Doubling Season and
//     Vorinclex all apply. The CR 614 ENTRY pipeline is not involved
//     and must not be: the land is already on the battlefield, so
//     these are placed counters, not entry counters (#1120's
//     distinction), and a Corpsejack Menace sees them for that reason.
//  4. The DELAYED RETURN is a CR 603.7 delayed triggered ability keyed
//     on the OBJECT (`DelayedTrigger.On` + `AppliesTo`, #663), not a
//     step. See `scheduleEarthbendReturnLocked`.
//
// # The duration is indefinite, pinned to the object
//
// Earthbend states no duration, so CR 611.2a says the animation lasts
// until the game ends — but CR 400.7 says the permanent that comes
// back after it dies is a NEW OBJECT, and the effect names the old
// one. That is `IndefiniteDuration()` through `Game.PinnedTo`: the
// affected set is keyed on {instance, battlefield-entry stamp} so the
// effect stops applying the instant the land leaves, and the pin
// makes `durationExpiredLocked` drop the registry entry at the next
// sweep rather than leaving a husk per earthbend for the rest of the
// game.
//
// The animation is ONE data record (ADR 0041 phase 3, #1497) with a mod
// in each of its three layers, so ONE timestamp and ONE duration are a
// fact about the data: that is what makes it one continuous effect for
// CR 613.7, and "the haste and the animation end together" something
// that cannot come apart. Being data, it is carried by a restore point
// rather than blocking one.

// EarthbendForEffect takes the earthbend keyword action on behalf of
// `actor`: it opens the CR 614 window on the action and, once the
// window settles, animates `land`, puts the settled count of +1/+1
// counters on it, and schedules the return-on-death-or-exile.
//
// `source` is the card whose effect is earthbending — the resolving
// spell, the Saga chapter's enchantment, the activated ability's
// permanent. It is the attribution on the continuous effects and the
// `SourceCardID` of the delayed trigger.
//
// `land` is the TARGET. Target legality (a land, controlled by the
// actor) belongs to the card's own `TargetSpec` and the CR 608.2b
// re-check, which is why this takes an instance ID and not a
// predicate; a land that has left the battlefield by the time the
// action is taken is a silent no-op, which is what CR 608.2b's
// re-check already decided for the spell that named it.
//
// It can PAUSE — a window with two count replacements in it queues
// the CR 616 ordering prompt and returns with NOTHING done; the
// resume earthbends when the prompt is answered, through the same
// `applyResolvedKeywordActionLocked` the unpaused path runs.
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) EarthbendForEffect(actor, source, land uuid.UUID, n int) error {
	return g.EarthbendThenForEffect(actor, source, land, n, nil)
}

// EarthbendThenForEffect is EarthbendForEffect with the rest of the
// sentence: Earthshape's "Earthbend 3. Then each creature you control
// with power less than or equal to that land's power gains hexproof".
//
// `then` runs once the earthbend has actually finished — after the
// counters have LANDED, which on a board with two different counter
// replacements (a Doubling Season beside a Hardened Scales) is when
// the CR 616 ordering prompt is answered, not when this returns
// (#1282). It also runs when the earthbend did nothing (replaced away,
// or its land has left): the sentence after "then" is not conditional
// on the keyword action having happened.
//
// A nil land is a no-op that still runs `then`, for the same reason.
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) EarthbendThenForEffect(actor, source, land uuid.UUID, n int, then func(g *Game) error) error {
	if land == uuid.Nil {
		if then != nil {
			return then(g)
		}
		return nil
	}
	_, err := g.runKeywordActionLocked(&ReplacementEvent{
		Kind:               RepEventKeywordAction,
		Actor:              actor,
		Source:             source,
		KeywordAction:      KeywordActionEarthbend,
		KeywordActionCount: n,
		keywordAction:      &keywordActionTail{land: land, then: then},
	})
	return err
}

// applyEarthbendLocked is the settled action: the four parts, in
// printed order except that the counters go LAST.
//
// The counters are last because they are the only part that can pause
// (a CR 616 ordering prompt between two counter replacements), and
// the placement returns nil on that pause with the counters owed to
// the resume. Everything the earthbend owes has to be registered
// before then, or a paused Doubling Season prompt would leave a land
// that is not a creature and has no delayed return — the ordering
// argument `enterBattlefieldThroughPipelineLocked`'s tail makes,
// applied to a verb.
//
// `then` is the rest of the sentence, and it rides the placement's
// continuation (#1282) rather than running on the next line, so
// "that land's power" is read once the counters are really there. It
// runs exactly once on every path: land gone, earthbend 0, and the
// counters landed (inline or from the resume).
//
// Printed order is not otherwise observable: state-based actions run
// when a player would receive priority (CR 704.3), and no player does
// in the middle of one resolution, so the 0/0 the animation makes
// cannot die before the counters land.
//
// Caller must hold g.mu.
func (g *Game) applyEarthbendLocked(actor, source, land uuid.UUID, n int, then func(g *Game) error) error {
	rest := func(g *Game) error {
		if then == nil {
			return nil
		}
		return then(g)
	}
	c, ok := g.battlefieldCardLocked(land)
	if !ok {
		// CR 608.2b already had its say about the target; a land that
		// left between the re-check and here is a no-op, not an error.
		return rest(g)
	}
	stamp := c.EnteredBattlefieldAt
	g.animateEarthbentLandLocked(source, land, stamp)
	g.scheduleEarthbendReturnLocked(actor, source, land, stamp)
	if n <= 0 {
		// "Earthbend 0" is a real instruction — Rockalanche with no
		// Forests — and it is not a no-op: the land is a 0/0 creature
		// with no counters, so the toughness SBA kills it and the
		// delayed return above hands it back tapped. Placing zero
		// counters is the one part that does not happen.
		return rest(g)
	}
	// The animation above only REGISTERED its layer effects; the
	// cached characteristic still says "Land" and not "Land Creature"
	// until something recomputes. A counter replacement that asks "is
	// this a creature you control" — Hardened Scales, Branching
	// Evolution, Corpsejack Menace — reads that cache, so without the
	// recompute it never saw an earthbend at all (found by #1282's
	// real-card proof test). The recompute is a no-op when nothing is
	// stale.
	g.RecomputeLayersIfStaleLocked()
	return g.AddCounterByThenForEffect(actor, land, CounterPlusOne, n, func(g *Game, _ int) error {
		return rest(g)
	})
}

// earthbendLabel is the attribution the continuous effect and
// the delayed trigger share. One string so a stall dump, the layer
// census and a test failure all name the same verb.
const earthbendLabel = "earthbend"

// animateEarthbentLandLocked registers the continuous effect that
// makes up "becomes a 0/0 creature with haste that's still a land":
// ONE data record (ADR 0041 phase 3, #1497) with a mod in each of
// three layers — layer 4 (add Creature), layer 6 (haste) and layer 7b
// (base P/T 0/0).
//
// One record is one timestamp and one duration across all three
// (CR 613.7, CR 611.2a + CR 400.7 — see the file comment), and one
// affected set, pinned to {instance, entry stamp} so a land that
// leaves and returns is correctly a different object (CR 611.2c,
// CR 400.7). It is data, so the animation no longer keeps its table
// off the restore path for as long as the land lives. (The delayed
// return below still does, until ADR 0041 phase 3's tier 2.)
//
// The layers, and why each is where it is:
//
//   - Layer 4 (CR 613.1d): Creature is ADDED. Land is not touched —
//     "that's still a land" is the reminder text spelling out what
//     adding a type already means, and it is load-bearing: the land
//     keeps its mana ability and still counts for landfall, land
//     counts and "lands you control".
//   - Layer 7b (CR 613.4b): base power and toughness become 0/0. The
//     +1/+1 counters apply at 7d over the top, which is why an
//     earthbend 4 is a 4/4 and why a -1/-1 counter from elsewhere
//     still shrinks it. 7b and not 7a: 7a is for characteristic-
//     defining abilities, and this is an effect from a resolved spell
//     or ability setting a specific value. Setting P/T here is also
//     what makes the land's toughness KNOWN (`Characteristic.PTDefined`,
//     #690), which is what lets the CR 704.5f state-based action see a
//     0/0 with no counters and kill it — the whole point of earthbend 0.
//   - Layer 6 (CR 613.1f): haste. The grant has no stated duration of
//     its own, so it takes the animation's: the land can attack the
//     turn it was earthbent (CR 302.6 would otherwise forbid it,
//     because it has not been controlled continuously as a CREATURE
//     since the turn began — #537) and it can still attack four turns
//     later, which an until-end-of-turn grant would not give it.
//
// Registering it TWICE on the same object is harmless and is what a
// second earthbend does: each mod is idempotent (the type and the
// keyword are appended only when absent, the base P/T is set to the
// same 0/0), and the layer pass sorts the two records by timestamp
// with the same result either way.
//
// Caller must hold g.mu.
func (g *Game) animateEarthbentLandLocked(source, land uuid.UUID, stamp int64) {
	mods := []Mod{AddTypesMod("Creature"), AddKeywordsMod("haste")}
	mods = append(mods, SetBasePTMods(0, 0)...)
	g.RegisterScopedEffectForEffect(source,
		[]AffectedObject{{ID: land, EnteredAt: stamp}}, mods,
		g.PinnedTo(IndefiniteDuration(), land),
		earthbendLabel+" — becomes a 0/0 creature with haste that's still a land")
}

// earthbendReturnNamespace seeds the deterministic ID every earthbend
// return trigger carries. See scheduleEarthbendReturnLocked.
var earthbendReturnNamespace = uuid.MustParse("e4a17be4-0000-4000-8000-000000001178")

// earthbendReturnTriggerID is the delayed trigger's identity: the
// OBJECT it watches, {instance, battlefield-entry stamp}, hashed into
// a UUID. Two earthbends on the same object compute the same ID; the
// same land earthbent again after it has died and come back computes
// a different one, because the stamp is re-minted on every entry
// (CR 400.7).
//
// A deterministic ID rather than a new field on DelayedTrigger: the
// queue is plain data that clones and snapshots by value, and this
// needs no schema change to answer "is one already watching this
// object".
func earthbendReturnTriggerID(land uuid.UUID, enteredAt int64) uuid.UUID {
	key := make([]byte, 0, len(land)+8)
	key = append(key, land[:]...)
	for shift := 56; shift >= 0; shift -= 8 {
		key = append(key, byte(enteredAt>>shift))
	}
	return uuid.NewSHA1(earthbendReturnNamespace, key)
}

// scheduleEarthbendReturnLocked creates "when it dies or is exiled,
// return it to the battlefield tapped" as a CR 603.7 delayed
// triggered ability keyed on the OBJECT.
//
// # Why a delayed trigger and not a card hook or a replacement
//
// It is a delayed TRIGGERED ability, so it uses the stack (CR 603.7b)
// and every player gets a response window before the land comes back;
// and it belongs to no permanent, because the permanent it is about
// is the thing that just left. `DelayedTrigger.On` (#663) is exactly
// that shape with an EVENT condition instead of a step: `EventLTB`,
// narrowed by `earthbendReturnMatches` to this object going to a
// graveyard or to exile. It fires ONCE and ceases to exist, which is
// CR 603.7b for free — `fireEventDelayedTriggersLocked` removes a
// matched trigger from the queue before it dispatches.
//
// It is not a replacement effect: the land really does die, so
// "whenever a creature you control dies" payoffs, Blood Artist and a
// graveyard count all see it. Only what happens NEXT is added.
//
// # Why the NewZone test is the right reading of "dies or is exiled"
//
// `EventLTB.NewZone` is where the card ACTUALLY went, after the CR 614
// window on the move settled. So a Rest in Peace-style replacement
// that exiles the land instead of letting it die still satisfies the
// trigger — and it satisfies the printed text, which names both
// destinations. A bounce to hand or a tuck into a library satisfies
// neither, and the trigger stays queued until the pin sweeps it.
//
// # Why two earthbends do not stack two returns
//
// The trigger's ID is derived from the object it watches, and a
// second earthbend on the same object finds it already queued and
// adds nothing. In paper there really would be two delayed abilities,
// both triggering; the first returns the land, the second finds a
// card that is no longer where it was left — a NEW OBJECT on the
// battlefield (CR 400.7) — and does nothing. One queue entry is the
// same game state with less bookkeeping.
//
// DECLARED DIVERGENCE, and it is one line wide: if the first return
// were countered (a "counter target triggered ability" effect), paper
// would still have the second and this would not. Nothing in the
// catalog counters a triggered ability — it is an open seam
// ("Ability-copy" and its neighbours), and when one lands this is the
// place that has to grow a real per-instance queue.
//
// # The duration is the garbage collector
//
// Indefinite, pinned to the object. Nothing ENDS an earthbend return
// on a turn boundary — the land can sit there for ten turns and still
// come back when it dies — but a land that leaves the battlefield any
// OTHER way (bounced, tucked, shuffled away) can never satisfy the
// trigger again, because the returning card is a new object. The pin
// is what drops the dead entry at the next sweep instead of carrying
// it for the rest of the game.
//
// Caller must hold g.mu.
func (g *Game) scheduleEarthbendReturnLocked(actor, source, land uuid.UUID, stamp int64) {
	id := earthbendReturnTriggerID(land, stamp)
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.ID == id {
			return
		}
	}
	d := g.PinnedTo(IndefiniteDuration(), land)
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		ID:           id,
		Controller:   actor,
		SourceCardID: source,
		Label:        earthbendLabel + " — when it dies or is exiled, return it to the battlefield tapped",
		On:           []EventKind{EventLTB},
		AppliesTo:    earthbendReturnMatches,
		Cards:        []uuid.UUID{land},
		Duration:     &d,
		Effect:       returnEarthbentLandTapped,
	})
}

// earthbendReturnMatches is the event condition: this object leaving
// the battlefield FOR a graveyard or for exile.
//
// A package-level func capturing nothing, on the contract
// DelayedTrigger.AppliesTo documents — the object it is about rides on
// the trigger's own Cards payload, which cloneDelayedTrigger
// reallocates.
func earthbendReturnMatches(ev Event, dt *DelayedTrigger, _ *Game) bool {
	if ev.Kind != EventLTB || len(dt.Cards) == 0 || ev.CardID != dt.Cards[0] {
		return false
	}
	return ev.NewZone == ZoneGraveyard || ev.NewZone == ZoneExile
}

// returnEarthbentLandTapped is what the fired trigger does: put the
// card back onto the battlefield tapped, under its OWNER's control
// (CR 400.3 — the printed text names no controller, and a land
// somebody stole and killed goes home).
//
// Package-level and capturing nothing: the card rides on the item's
// Targets, the way every delayed trigger's payload does.
//
// A card that is no longer in a graveyard or in exile is skipped
// SILENTLY — CR 608.2b's posture and `returnExiledCardsToOwners`'s
// contract — which is why `ErrCardNotFound` is swallowed here and
// nothing else is. That covers a land somebody reanimated out of the
// graveyard first, a land whose graveyard was exiled wholesale, and
// (if a future per-instance queue ever schedules two) the second
// trigger of a doubled return. An instruction that legally does
// nothing is a resolution, not a card that threw: an EventEffectError
// here would fail a catalog soak over a board that is perfectly legal.
//
// Caller holds g.mu (it runs as a stack item's Effect).
func returnEarthbentLandTapped(g *Game, item *StackItem) error {
	for _, t := range item.Targets {
		if t.Kind != TargetCard || t.ID == uuid.Nil {
			continue
		}
		_, err := g.ReturnToBattlefieldForEffect(t.ID, uuid.Nil, true)
		if err != nil && !errors.Is(err, ErrCardNotFound) {
			return err
		}
	}
	return nil
}
