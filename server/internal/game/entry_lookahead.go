package game

import (
	"github.com/google/uuid"
)

// entry_lookahead.go — the CR 614.12 look-ahead (ADR 0109 §10, owner
// decision 1, #1556).
//
//	CR 614.12  "To determine which replacement effects apply and how
//	            they apply, check the characteristics of the permanent
//	            as it would exist on the battlefield, taking into
//	            account replacement effects that have already modified
//	            how it enters the battlefield (see rule 616.1),
//	            continuous effects from the permanent's own static
//	            abilities that would apply to it once it's on the
//	            battlefield, and continuous effects that already exist
//	            and would apply to the permanent."
//
// The entry window runs BEFORE the permanent is on the battlefield, so
// the layer pass has never computed it: the card in its source zone has
// only its printed keywords. A creature entering under Rhythm of the
// Wild ("Nontoken creatures you control have riot") has riot as it
// enters, and one entering under Dress Down ("Creatures lose all
// abilities") has none, but neither shows on the card in its hand or on
// the stack.
//
// entryLookAheadLocked answers it with a DRY RUN of the layer pass: the
// battlefield as it is, plus the entering card as the permanent it would
// be — under its would-be controller (ev.Actor), as the copy it was told
// to enter as (CR 616.1c, ADR 0102 decision 6), and with the records an
// effect pinned to it as a spell, which CR 400.7a carries onto the
// permanent. The pass runs over a COPY of the battlefield slice, so
// every write it makes lands on the copy; the real slice, its cards and
// their layer caches are untouched, and nothing is emitted. The dry
// permanent has an ID of its own, so no instance ID is ever in two
// zones (ADR 0094).
//
// It reads keywords, counting instances (CR 702.136b), and through them
// ability removal, and nothing else. A face-down entry has no abilities
// (CR 708.2a) and is never looked at.

// entryLookAheadNamespace seeds the ID the dry permanent runs under.
var entryLookAheadNamespace = uuid.MustParse("1556e0a1-7a11-4c0a-8d0e-000000001556")

// entryLookAhead is what the look-ahead reports about the entering
// permanent: how many instances of each entry keyword it would have.
type entryLookAhead struct {
	riot    int
	unleash int
}

// entryLookAheadCache is one entry event's memo of its look-ahead,
// keyed on everything that can change the answer while the window is
// open: the layer version (a static entered or left, a record was
// registered), the would-be controller (an entry_controller answer), and
// the copy the card enters as.
//
// The layer version is not monotone across an undo, but the key still
// holds there: a paused entry's event is copied by value into the
// snapshot (cloneReplacementResume), so the snapshot's copy carries a
// cache computed at or before the snapshot's version, and an undo
// restores the game at that version plus one (restoreFromLocked), which
// matches none. The cache itself is never written in place.
type entryLookAheadCache struct {
	version uint64
	actor   uuid.UUID
	copyOf  *PrintedValues
	result  entryLookAhead
}

// entryLookAheadLocked is the look-ahead for the permanent `ev` would
// put onto the battlefield. The zero value — no entry keywords — for any
// event that is not a battlefield entry of a card not already there, and
// for a face-down entry.
//
// Caller must hold g.mu (write): the dry pass assigns layer caches on the
// copied battlefield.
func (g *Game) entryLookAheadLocked(ev *ReplacementEvent) entryLookAhead {
	if ev == nil || ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.CardID == uuid.Nil ||
		ev.FaceDown != FaceDownNone || g.Battlefield == nil || g.Battlefield.Contains(ev.CardID) {
		return entryLookAhead{}
	}
	version := g.layerVersion.Load()
	if c := ev.lookAhead; c != nil && c.version == version && c.actor == ev.Actor && c.copyOf == ev.EntersAsCopyOf {
		return c.result
	}
	var result entryLookAhead
	if ch, ok := g.entryCharacteristicsLocked(ev); ok {
		for _, a := range ch.Abilities {
			switch a {
			case KeywordRiot:
				result.riot++
			case KeywordUnleash:
				result.unleash++
			}
		}
	}
	ev.lookAhead = &entryLookAheadCache{version: version, actor: ev.Actor, copyOf: ev.EntersAsCopyOf, result: result}
	return result
}

// entryCharacteristicsLocked is the dry run itself: the entering
// permanent's effective characteristics as they would be on the
// battlefield. ok is false when the entering card cannot be found.
//
// Caller must hold g.mu (write).
func (g *Game) entryCharacteristicsLocked(ev *ReplacementEvent) (Characteristic, bool) {
	entering, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return Characteristic{}, false
	}
	onStack := g.Stack != nil && g.Stack.Contains(ev.CardID)
	stackEpoch := entering.ObjectEpoch

	perm := entering
	// CR 614.12 / CR 616.1c: a copy it has already been told to enter
	// as is what it would be on the battlefield (ADR 0102 decision 6).
	if ev.EntersAsCopyOf != nil {
		perm.applyCopy(*ev.EntersAsCopyOf, Card{})
	}
	// The would-be controller: the entry's actor, which an
	// entry_controller answer rewrites (ADR 0102 decision 2).
	controller := ev.Actor
	if controller == uuid.Nil {
		controller = perm.Controller
	}
	if controller == uuid.Nil {
		controller = perm.Owner
	}
	perm.Controller = controller
	perm.BaseController = controller
	// The new object it would be (CR 400.7): the epoch MoveCard would
	// give it, an entry stamp later than every permanent already there
	// (CR 613.7d), and no layer cache or stack grants of its own — the
	// pass computes the first, and the records behind the second are
	// carried below.
	//
	// It runs under an ID of its own, derived from the entering card's,
	// because the card is still in its source zone and the card index
	// (ADR 0094) holds that one instance ID is in one zone. Nothing on
	// the battlefield can name the permanent before it exists, so the
	// only things that look it up by ID are the records carried below,
	// which are pinned to the look-ahead ID.
	perm.InstanceID = uuid.NewSHA1(entryLookAheadNamespace, entering.InstanceID[:])
	perm.ObjectEpoch = stackEpoch + 1
	perm.EnteredBattlefieldAt = g.lookAheadEntryStampLocked()
	perm.SummonedThisTurn = true
	perm.effective = nil
	perm.stackGranted = nil
	perm.Tapped = ev.EntersTapped

	// CR 400.7a: an effect that changed the SPELL's characteristics
	// continues to apply to the permanent it becomes. Those are the
	// records pinned to the spell (ADR 0104, ADR 0107 §3); the landing
	// re-pins them to the permanent (repinSpellControlLocked), and the
	// dry run applies them to it the same way, at their own timestamps.
	var extra []ContinuousEffect
	if onStack {
		var carried []ScopedEffect
		for _, e := range g.ScopedEffects {
			if !pinsStackObject(e.Affected, entering.InstanceID, stackEpoch) {
				continue
			}
			e.Affected = []AffectedObject{PinObjectByEpoch(perm.InstanceID, perm.ObjectEpoch)}
			carried = append(carried, e)
		}
		extra = adaptScopedEffects(carried)
	}

	orig := g.Battlefield.Cards
	dry := make([]Card, len(orig), len(orig)+1)
	copy(dry, orig)
	dry = append(dry, perm)
	g.Battlefield.Cards = dry
	defer func() { g.Battlefield.Cards = orig }()
	g.layerPassWithLocked(extra)
	ch := dry[len(dry)-1].effective
	if ch == nil {
		return Characteristic{}, false
	}
	return *ch, true
}

// lookAheadEntryStampLocked is the entry stamp the dry permanent runs
// under: one past the latest timestamp on the board — every permanent's
// layer timestamp and every resolved record's — so it sorts after all of
// them, as a permanent entering now does (CR 613.7d). It is derived
// rather than read off the clock, because the look-ahead writes nothing
// and must not advance anything, a test clock included.
//
// Caller must hold g.mu.
func (g *Game) lookAheadEntryStampLocked() int64 {
	var latest int64
	for i := range g.Battlefield.Cards {
		latest = max(latest, g.Battlefield.Cards[i].layerTimestamp())
	}
	for i := range g.ScopedEffects {
		latest = max(latest, g.ScopedEffects[i].Timestamp)
	}
	return latest + 1
}
