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
// "you may pay 2 life", Clone's "choose what to copy", a CR 614.10
// "may". Two entry sites — the land play and stack resolution — carried
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
}

// runEntryTailLocked runs a settled entry's continuation exactly once.
// The continuation is cleared before it runs, so a tail that re-enters
// the pipeline on the same event cannot run itself twice.
//
// Caller must hold g.mu.
func (g *Game) runEntryTailLocked(ev *ReplacementEvent, entered uuid.UUID) error {
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
	if out == nil || out.Canceled || out.NewZone != ZoneBattlefield {
		// CR 614.10 with a null replacement, or a replacement that sent
		// the card somewhere else. There is no generic "put it wherever
		// the pipeline said" helper for these sources, so a redirect is
		// treated as a cancel rather than guessed at. The caller's
		// continuation still runs — see runEntryTailLocked.
		return uuid.Nil, nil
	}
	return g.executeEntryToBattlefieldLocked(out)
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
		c.LostLastCounter = false
		c.KnownBy = nil
		c.DamageMarked = 0
		c.MarkedLethalByDeathtouch = false
		c.RegenerationShields = 0
		c.AttackingTarget = uuid.Nil
		c.BlockingTarget = uuid.Nil
		c.GoadedBy = uuid.Nil
		c.ClearFaceDown()
		c.BattleX = 0
		c.BattleY = 0
		c.EnteredBattlefieldAt = 0
		c.SummonedThisTurn = false
		c.NamedTribe = ""
		c.ChosenColor = ""
		c.ClassLevel = 0
		c.Solved = false
		c.effective = nil
		return newID
	}
	return uuid.Nil
}
