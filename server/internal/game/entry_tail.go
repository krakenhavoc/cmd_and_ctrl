package game

import (
	"errors"

	"github.com/google/uuid"
)

// entry_tail.go — the ENTRY half of what zone_route.go is for exits
// (#478).
//
// Why it exists. A battlefield entry runs the CR 614 pipeline before
// the card leaves its old zone, and that pipeline can PAUSE: a CR 616
// ordering prompt between two enters-tapped replacements, a shockland's
// "you may pay 2 life", Clone's "choose what to copy", a
// "may" replacement. Two entry sites — the land play and stack resolution — carried
// `entryResumable` and were finished from
// executeEntryToBattlefieldLocked when the answer arrived. Every OTHER
// entry site dropped the move on a pause, and said so in a comment.
//
// The comment's reasoning was right about the cost and wrong about the
// conclusion. Resuming an entry generically would finish the MOVE while
// skipping what the effect that started it still owed — the library
// shuffle and EventSearchLibrary for a search, the caller's `Then`, and
// the new object identity (CR 400.7) an exile return mints. So the
// answer is not "don't resume"; it is to carry those duties across the
// pause the way a paused EXIT already carries its destination, its
// to-the-bottom instruction and its continuation on zoneRoute.
//
// That is this file. `entryTail` rides on the ReplacementEvent, inside
// the same `replacementResume` frame every other paused event uses, and
// the SAME resume finishes it: applyResolvedReplacementEventLocked's
// entry branch → executeEntryToBattlefieldLocked. There is no second
// entry resume, and the inline path goes through the same finisher, so
// a fetched permanent cannot enter differently depending on whether
// anybody happened to be asked a question.
//
// What a paused entry costs the table is one action, exactly as a
// paused exit does: the library is shuffled, the search event fires and
// the caller's continuation runs when the prompt is answered rather
// than on the line after the fetch.

// entryTail is what the effect that asked for a battlefield ENTRY still
// owes once the CR 614 pipeline settles — the entry-side twin of
// zoneRoute's per-destination bookkeeping and `then`.
//
// Unexported engine plumbing; the catalog never sets or reads one.
type entryTail struct {
	// newObject mints a fresh InstanceID and strips the old object's
	// battlefield state on arrival. CR 400.7: a card that changes
	// zones becomes a NEW OBJECT with no memory of the old one, and
	// the exile return is the entry where that is observable — it is
	// what makes a blink a removal answer (counters fall off, a
	// stolen creature goes home) as well as an ETB engine.
	//
	// Only the exile return sets it. Every other entry keeps the
	// instance ID the card had in its source zone, which is what
	// callers holding that ID expect.
	newObject bool

	// then is the rest of whatever asked for the entry, run once the
	// entry reaches a TERMINAL outcome — it entered, the window
	// cancelled or redirected it, it was refused, or its prompt was
	// taken away. A PAUSE is not terminal: the resume reaches the tail
	// later, through executeEntryToBattlefieldLocked.
	//
	// `entered` is the permanent's ID on the battlefield — the NEW one
	// when newObject minted it — and uuid.Nil when nothing entered. A
	// caller sequencing a batch through here has to be told even when
	// the answer is "nothing happened", or it waits forever; the same
	// contract runRouteTailLocked, runLifeTailLocked and
	// runDamageTailLocked hold.
	//
	// It takes the live *Game rather than capturing one, on the undo-
	// safety contract every continuation in the engine follows. It runs
	// with g.mu held, so it may start the next entry or queue the next
	// prompt itself.
	//
	// Cleared THROUGH the pointer as it runs (runEntryTailLocked),
	// which is why cloneReplacementResume gives an undo snapshot its
	// own copy of the tail.
	then func(g *Game, entered uuid.UUID) error

	// batch is the simultaneous entry this event is one card of
	// (entry_batch.go, #1322), and batchIndex the card's place in it.
	// When set, a settled window does NOT land the card: the resume
	// hands the settled event back to the batch, which walks on to the
	// next card's window and lands them all together once the last one
	// settles. Detached through the pointer as it is used
	// (ReplacementEvent.takeEntryBatch), so exactly one path can hand
	// the event back; cloneReplacementResume gives an undo snapshot its
	// own copy of the batch.
	//
	// Never set together with `then`: a batch's continuation is the
	// batch's own, run once for the whole entry.
	batch      *entryBatch
	batchIndex int
}

// runEntryTailLocked runs a settled entry's continuation exactly once.
// The continuation is cleared before it runs, so a tail that re-enters
// the pipeline on the same event cannot run itself twice.
//
// For one card of a simultaneous entry (entryTail.batch) the
// "continuation" is the rest of the batch. This function is reached
// only on a terminal outcome where nothing entered — a cancel, a
// dropped or pruned prompt — because a settled entry of a batch card
// is intercepted before it can land (applyResolvedReplacementEventLocked),
// so the batch is told this card is not part of the entry and walks on.
//
// Caller must hold g.mu.
func (g *Game) runEntryTailLocked(ev *ReplacementEvent, entered uuid.UUID) error {
	if b, i, ok := ev.takeEntryBatch(); ok {
		return g.resumeEntryBatchLocked(b, i, nil)
	}
	if ev == nil || ev.entryTail == nil || ev.entryTail.then == nil {
		return nil
	}
	then := ev.entryTail.then
	ev.entryTail.then = nil
	return then(g, entered)
}

// enterBattlefieldThroughPipelineLocked is the ONE effect-side entry
// primitive: it opens the CR 614 window for a move of ev.CardID onto
// the battlefield and, once the pipeline settles, performs the move
// through executeEntryToBattlefieldLocked.
//
// routeCardToZoneLocked's shape, and deliberately so — this is the
// entry half of the same contract:
//
//   - When the pipeline queues a player prompt NOTHING has moved: the
//     card is still in its old zone, no event has been emitted, and the
//     entry completes from applyResolvedReplacementEventLocked when the
//     player answers. `entered` is uuid.Nil until then.
//   - A caller with something to do AFTER the entry hands it over as
//     ev.entryTail.then rather than writing it on the next line: the
//     continuation runs from the landing either way, inline when
//     nothing paused and from the resume when something did.
//   - Every exit of this function but the PAUSE is terminal for the
//     entry, so the tail runs from all of them.
//
// `entered` is the permanent's ID, uuid.Nil when nothing entered. The
// exit primitive returns its `paused` bool because one caller
// (ExileTopFaceDownForEffect) must not re-read the top of a library it
// did not move; no entry caller has that hazard, so the pause is not
// reported separately from "nothing entered".
//
// The event must be built by the caller (it is the caller that knows
// the actor, the source zone and the effect's own "enters tapped"
// clause) with entryResumable set — an entry that reaches here can
// pause, and an event that pauses with no resume is the bug this
// closes.
//
// Caller must hold g.mu.
func (g *Game) enterBattlefieldThroughPipelineLocked(ev *ReplacementEvent) (entered uuid.UUID, err error) {
	paused := false
	defer func() {
		if paused {
			return
		}
		// The landing path has already run the tail through this same
		// pointer by the time this fires, and runEntryTailLocked clears
		// the continuation as it runs, so it cannot run twice.
		if tailErr := g.runEntryTailLocked(ev, uuid.Nil); tailErr != nil && err == nil {
			err = tailErr
		}
	}()
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// A prompt is queued and the resume owns the entry now. The
		// event's tracking map entry deliberately survives: the resume
		// re-enters the apply-loop with the same ev.ID.
		paused = true
		return uuid.Nil, nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return uuid.Nil, err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		g.restoreEntryControllerLocked(ev)
		// CR 614.10 with a null replacement. The caller's continuation
		// still runs — see runEntryTailLocked.
		return uuid.Nil, nil
	}
	// A replacement that sent the card somewhere else (Mox Diamond's
	// graveyard) is finished there by the same finisher, which moves
	// it through moveRedirectedEntryLocked (ADR 0098 Decision 4). It
	// used to be treated as a cancel, leaving the card where it was.
	return g.executeEntryToBattlefieldLocked(out)
}

// moveRedirectedEntryLocked finishes an entry whose CR 614 window
// settled somewhere other than the battlefield — "If you don't, put it
// into its owner's graveyard" (Mox Diamond, Heart of Yavimaya, Lotus
// Vale), and whatever a later replacement then did to that modified
// event (CR 616.2: Rest in Peace's "exile it instead", CR 903.9).
//
// ADR 0098 Decision 4. Before it, three entry finishers treated such a
// redirect as a CANCEL (the card stayed where it was — for a resolving
// permanent spell, on the stack with its record already gone), and the
// stack-resolution and land-play sites did not check at all and put the
// card onto the battlefield anyway.
//
// The settled entry is turned into a settled EXIT and moved by
// executeZoneRouteLocked, the mover every routed exit uses, from
// whichever zone the card is in. It does NOT open a second window: the
// replacements that apply to the move already applied to this event
// while it was being replaced, and running them again would, for one,
// ask a commander's owner CR 903.9's question twice.
//
//   - A created token (not in any zone yet) simply ceases to be: a
//     token in a zone other than the battlefield ceases to exist
//     (CR 111.7).
//   - A card that has left the zone the window opened over is left
//     alone, as landEntryLocked leaves it.
//   - A destination that is the zone the card is already in (a Mox
//     Diamond reanimated with no land to discard) is nothing to do —
//     the sandbox mover's reading of the same case.
//   - A land PLAY that is redirected still spends the land drop
//     (CR 116.2a, 305.2; ADR 0098 owner decision 7): the play is the
//     special action, and it was taken.
//
// The caller runs the entry tail, with uuid.Nil.
//
// Caller must hold g.mu.
func (g *Game) moveRedirectedEntryLocked(ev *ReplacementEvent) error {
	src := g.findCardZoneLocked(ev.CardID)
	if src == nil {
		g.dropEnteringTokenLocked(ev.CardID)
		return nil
	}
	if src.Kind != ev.OldZone || src.Kind == ZoneBattlefield {
		return nil
	}
	// ADR 0102: the land drop is the player's who played it, and an
	// entry-controller effect's re-stamp does not follow the card
	// anywhere but the battlefield.
	landPlayer := ev.landPlayer
	if landPlayer == uuid.Nil {
		landPlayer = ev.Actor
	}
	if ev.landPlay && landPlayer != uuid.Nil {
		if g.LandsPlayedThisTurn == nil {
			g.LandsPlayedThisTurn = make(map[uuid.UUID]int)
		}
		g.LandsPlayedThisTurn[landPlayer]++
	}
	g.restoreEntryControllerLocked(ev)
	dst, _, err := g.routeDestinationLocked(ev.CardID, ev.NewZone, ev.NewZoneOwner)
	if err != nil {
		return err
	}
	if dst == src {
		return nil
	}
	ev.zoneRoute = &zoneRoute{
		CardID:   ev.CardID,
		Dst:      ev.NewZone,
		DstOwner: ev.NewZoneOwner,
		Actor:    ev.Actor,
		Cause:    g.resolutionCauseLocked(),
	}
	return g.executeZoneRouteLocked(ev)
}

// resetAsNewObjectLocked gives the battlefield card `oldID` a fresh
// instance ID and clears every scrap of the old object's battlefield
// state — CR 400.7's "a card that changes zones becomes a new object".
//
// The exile return is the one entry that does this, and it is here
// rather than inline in that entry point because the entry now settles
// in the shared finisher, on both the inline and the resumed path.
//
// The random-source ordinal is carried over deliberately: a blink is a
// new object but not a new die (see sourceOrdinals).
//
// Returns the new ID, or uuid.Nil when the card is not on the
// battlefield.
//
// Caller must hold g.mu.
func (g *Game) resetAsNewObjectLocked(oldID uuid.UUID) uuid.UUID {
	newID := uuid.New()
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != oldID {
			continue
		}
		if ordinal, ok := g.sourceOrdinals[oldID]; ok {
			g.sourceOrdinals[newID] = ordinal
		}
		c.InstanceID = newID
		c.Tapped = false
		c.NextUntapSkips = nil
		c.Counters = nil
		c.CounterStampedAt = nil
		c.LostLastCounter = false
		c.KnownBy = nil
		c.DamageMarked = 0
		c.MarkedLethalByDeathtouch = false
		c.RegenerationShields = 0
		c.AttackingTarget = uuid.Nil
		c.clearBlocking()
		c.Goads = nil
		c.ClearFaceDown()
		c.BattleX = 0
		c.BattleY = 0
		c.EnteredBattlefieldAt = 0
		c.FaceTurnedAt = 0
		c.SummonedThisTurn = false
		c.NamedTribe = ""
		c.ChosenColor = ""
		c.ChosenPlayer = uuid.Nil
		c.ChosenName = ""
		c.ChosenOption = ""
		c.ModesChosen = nil
		c.Provenance = CastProvenance{}
		c.ClassLevel = 0
		c.Solved = false
		c.Harnessed = false
		c.Monstrous = false
		c.Prepared = false
		c.effective = nil
		return newID
	}
	return uuid.Nil
}
