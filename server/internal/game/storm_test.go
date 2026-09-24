package game

import (
	"testing"

	"github.com/google/uuid"
)

// storm_test.go — CR 702.40a's count, and the record it is read from
// (ADR 0086). The catalog half — the Storm() keyword, the copies and
// the per-copy prompts — is in cards/effects/storm_test.go.
//
// Every test here is about one of the three readings of "copy it for
// each OTHER spell that WAS CAST BEFORE IT this turn", because each
// of them is a way the obvious implementation gets it wrong:
//
//	"each other spell"  — every player's, and never this one
//	"was cast"          — a countered spell was still cast
//	"before it"         — an index, which no counter can produce

// castInstantFor puts an instant in p's hand and casts it, returning
// its instance ID. Thin wrapper over castForTest for the tests below,
// which care about the ID and not the stack item.
func castInstantFor(t *testing.T, g *Game, p *Player) uuid.UUID {
	t.Helper()
	floatMana(p, "B")
	return castForTest(t, g, p, "{B}", CastSpellParams{}).ID
}

// --- "before it": the order ----------------------------------------

// TestTheTurnsCastOrderIsAnIndexPerSpell — the headline. Three spells
// in a turn have storm counts 0, 1, 2, and the count is a property of
// the SPELL, not of when it is asked.
func TestTheTurnsCastOrderIsAnIndexPerSpell(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	first := castInstantFor(t, g, me)
	second := castInstantFor(t, g, me)
	third := castInstantFor(t, g, me)

	g.mu.Lock()
	defer g.mu.Unlock()
	for i, id := range []uuid.UUID{first, second, third} {
		if got := g.SpellsCastBeforeThisTurn(id); got != i {
			t.Errorf("spell %d: storm count %d, want %d", i, got, i)
		}
	}
}

// TestASpellCastInResponseDoesNotMoveAnEarlierSpellsCount is why the
// count may be read at RESOLUTION (ADR 0086 Decision 2, CR 608.2h):
// the list is append-only, so nothing cast after a spell can change
// how many were cast before it. A storm trigger holds an instance ID
// and re-derives the number whenever it likes.
func TestASpellCastInResponseDoesNotMoveAnEarlierSpellsCount(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	castInstantFor(t, g, me)
	storm := castInstantFor(t, g, me)

	g.mu.Lock()
	atAnnounce := g.SpellsCastBeforeThisTurn(storm)
	g.mu.Unlock()

	// Two more spells go on the stack above it, as a response would.
	castInstantFor(t, g, me)
	castInstantFor(t, g, me)

	g.mu.Lock()
	defer g.mu.Unlock()
	atResolution := g.SpellsCastBeforeThisTurn(storm)
	if atAnnounce != 1 || atResolution != 1 {
		t.Fatalf("storm count moved: %d at announce, %d after two responses, want 1 and 1",
			atAnnounce, atResolution)
	}
}

// --- "each other spell": whose, and not this one --------------------

// TestTheCastOrderCountsEveryPlayersSpells — CR 702.40a says "each
// other spell", not "each other spell you cast". A per-player reading
// (Game.SpellsCastThisTurn) makes Brain Freeze a blank.
func TestTheCastOrderCountsEveryPlayersSpells(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]

	castInstantFor(t, g, me)
	castInstantFor(t, g, opp)
	mine := castInstantFor(t, g, me)

	g.mu.Lock()
	defer g.mu.Unlock()
	if got := g.SpellsCastBeforeThisTurn(mine); got != 2 {
		t.Errorf("storm count = %d, want 2 (an opponent's spell counts)", got)
	}
	// And the per-player tally still says what it always said, so
	// nothing that reads "your first spell this turn" moved.
	if got := g.CastTallyFor(me.ID).Total; got != 2 {
		t.Errorf("CastTallyFor(me).Total = %d, want 2", got)
	}
}

// TestALandPlayIsNotACast — CR 116.2a. Playing a land is a special
// action, takes CastSpell's land branch and emits no EventCast, so it
// never reaches the record.
func TestALandPlayIsNotACast(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	land := NewCard("Test Island", me.ID)
	land.TypeLine = "Basic Land — Island"
	land.Controller = me.ID
	me.Hand.PushTop(land)
	if err := g.CastSpell(me.ID, land.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("play the land: %v", err)
	}
	spell := castInstantFor(t, g, me)

	g.mu.Lock()
	defer g.mu.Unlock()
	if got := g.SpellsCastBeforeThisTurn(spell); got != 0 {
		t.Errorf("storm count = %d, want 0 (a land play is not a cast)", got)
	}
	if n := len(g.TurnTally.Casts); n != 1 {
		t.Errorf("cast order holds %d entries, want 1", n)
	}
}

// TestASpellCopyIsNotACast — CR 707.10. A copy is created, not cast;
// createSpellCopyLocked emits no EventCast, so a storm chain counts
// only the real casts, as printed. Without this the second Grapeshot
// of a turn would copy itself into an unbounded chain.
func TestASpellCopyIsNotACast(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	original := castInstantFor(t, g, me)

	g.mu.Lock()
	before := len(g.TurnTally.Casts)
	if err := g.CopySpellForEffect(original, me.ID, false, nil); err != nil {
		t.Fatalf("CopySpellForEffect: %v", err)
	}
	after := len(g.TurnTally.Casts)
	g.mu.Unlock()

	if after != before {
		t.Errorf("the copy joined the cast order: %d entries, want %d", after, before)
	}
	if g.Stack.Size() != 2 {
		t.Fatalf("the copy did not reach the stack: %d items", g.Stack.Size())
	}
}

// TestATriggerIsNotASpell — a triggered ability put on the stack is
// not cast (CR 603.3) and emits no EventCast. Storm's own trigger is
// the one that would otherwise inflate the next spell's count by one.
func TestATriggerIsNotASpell(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	const oracle = "storm-test-fromstack"

	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			FromStack: true,
			Watches:   []EventKind{EventCast},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "storm-test trigger",
					func(*Game, *StackItem) error { return nil })
			},
		}}
	})

	c := NewCard("Storm Test Spell", me.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{B}"
	c.OracleID = oracle
	c.Controller = me.ID
	me.Hand.PushTop(c)
	floatMana(me, "B")
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if len(g.PendingTriggers) == 0 && len(g.StackMeta) < 2 {
		t.Fatal("the FromStack trigger never fired — the fixture is wrong, not the count")
	}
	next := castInstantFor(t, g, me)

	g.mu.Lock()
	defer g.mu.Unlock()
	if got := g.SpellsCastBeforeThisTurn(next); got != 1 {
		t.Errorf("storm count = %d, want 1 (a trigger is not a spell)", got)
	}
}

// TestARecastSpellCountsFromItsLatestCast — Remand's shape. A spell
// bounced back to hand and cast again the same turn is two casts of
// one card, and a card keeps its instance ID across the round trip, so
// the record holds that ID twice. The question is about the LATER
// cast: taking the first index would undercount every recast, and the
// earlier cast's own trigger is long gone by then (CR 603.3 puts it
// above its own spell, so it resolves before anything can bounce
// that spell).
//
// The first cast counts toward the second: CR 400.7 makes the card a
// new object when it leaves the stack, so it really is "another
// spell that was cast before it".
func TestARecastSpellCountsFromItsLatestCast(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	bounced := castInstantFor(t, g, me)
	g.mu.Lock()
	err := g.ReturnSpellToHandForEffect(bounced)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ReturnSpellToHandForEffect: %v", err)
	}
	castInstantFor(t, g, me)
	floatMana(me, "B")
	if err := g.CastSpell(me.ID, bounced, CastSpellParams{}); err != nil {
		t.Fatalf("recast: %v", err)
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	if n := len(g.TurnTally.Casts); n != 3 {
		t.Fatalf("cast order holds %d entries, want 3 (the same card twice plus one other)", n)
	}
	if got := g.SpellsCastBeforeThisTurn(bounced); got != 2 {
		t.Errorf("storm count = %d, want 2 (the latest cast's index, not the first)", got)
	}
}

// --- "was cast": a countered spell still counts ---------------------

// TestACounteredSpellStaysInTheTurnsCastOrder — CR 702.40a counts
// spells that WERE CAST, whatever became of them. This is the reading
// an EventsThisTurn scan gets wrong: Thousand-Year Storm's helper
// looks each cast card up to test its type and drops one it cannot
// find, which is "weaker, never stronger" for that card and simply
// wrong for storm.
func TestACounteredSpellStaysInTheTurnsCastOrder(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	doomed := castInstantFor(t, g, me)
	g.mu.Lock()
	err := g.CounterTargetForEffect(doomed)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("CounterTargetForEffect: %v", err)
	}
	storm := castInstantFor(t, g, me)

	g.mu.Lock()
	defer g.mu.Unlock()
	if got := g.SpellsCastBeforeThisTurn(storm); got != 1 {
		t.Errorf("storm count = %d, want 1 (a countered spell was still cast)", got)
	}
}

// --- the announcement -----------------------------------------------

// TestStormCountAnnouncesItself — the count is the card, and nothing
// else in the log says it (ADR 0086 Decision 5).
func TestStormCountAnnouncesItself(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	castInstantFor(t, g, me)
	castInstantFor(t, g, me)
	storm := castInstantFor(t, g, me)

	var n int
	g.WithWriteLock(func() { n = g.StormCountForEffect(me.ID, storm) })
	if n != 2 {
		t.Fatalf("StormCountForEffect = %d, want 2", n)
	}

	var found *Event
	for i := range g.Events {
		if g.Events[i].Kind == EventStorm {
			found = &g.Events[i]
		}
	}
	if found == nil {
		t.Fatal("no EventStorm was emitted — the table is never told the count")
	}
	if found.Amount != 2 || found.CardID != storm || found.Actor != me.ID {
		t.Errorf("EventStorm = {amount %d, card %v, actor %v}, want {2, the storm spell, the caster}",
			found.Amount, found.CardID, found.Actor)
	}
}

// TestStormCountAnnouncesZero — a trigger that found nothing to copy
// is a different fact from a trigger that never fired, and the log is
// the only place that difference is visible.
func TestStormCountAnnouncesZero(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	first := castInstantFor(t, g, me)

	var n int
	g.WithWriteLock(func() { n = g.StormCountForEffect(me.ID, first) })
	if n != 0 {
		t.Fatalf("the turn's first spell has storm count %d, want 0", n)
	}
	for _, ev := range g.Events {
		if ev.Kind == EventStorm {
			return
		}
	}
	t.Error("no EventStorm for a count of zero")
}

// --- the record's lifetime ------------------------------------------

// TestTheTurnBoundaryClearsTheCastOrder — "this turn" (ADR 0059's
// real turn boundary, #1009), not "this game".
func TestTheTurnBoundaryClearsTheCastOrder(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	castInstantFor(t, g, g.Seats[0])

	g.mu.Lock()
	if len(g.TurnTally.Casts) != 1 {
		g.mu.Unlock()
		t.Fatalf("cast order holds %d entries before the turn ends, want 1", len(g.TurnTally.Casts))
	}
	g.mu.Unlock()

	for i := 0; g.Turn.ActiveSeat == 0 && i < 40; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == 0 {
		t.Fatal("never reached the next turn")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if n := len(g.TurnTally.Casts); n != 0 {
		t.Errorf("cast order survived the turn boundary: %d entries", n)
	}
}

// TestUndoRewindsTheTurnsCastOrder — an undo across a cast must take
// the cast out of the count, or the next storm spell copies itself
// for a spell that was never cast.
func TestUndoRewindsTheTurnsCastOrder(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	castInstantFor(t, g, me)
	snap := g.Clone()

	castInstantFor(t, g, me)
	storm := castInstantFor(t, g, me)
	g.mu.Lock()
	before := g.SpellsCastBeforeThisTurn(storm)
	g.mu.Unlock()
	if before != 2 {
		t.Fatalf("pre-undo storm count = %d, want 2", before)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })

	g.mu.Lock()
	defer g.mu.Unlock()
	if n := len(g.TurnTally.Casts); n != 1 {
		t.Errorf("undo left %d casts in the order, want 1", n)
	}
	if got := g.SpellsCastBeforeThisTurn(storm); got != 0 {
		t.Errorf("an undone cast still counts: storm count %d, want 0", got)
	}
}

// TestCloneDoesNotShareTheCastOrderBackingArray — cloneTurnTally
// copies the slice rather than aliasing it, and this is the
// arrangement in which that matters.
//
// Aliasing is invisible for an append-only list until a restore point
// is taken, undone past, and then played forward again: append reuses
// SPARE CAPACITY, so the replay writes into the same array an OLDER
// snapshot is still reading, one index past that snapshot's own
// length but inside the newer one's. Three casts is the smallest list
// with spare capacity (Go grows a nil slice 1, 2, 4), so the fourth
// cast lands in a cell the fifth would overwrite.
//
// Game.Events takes the other decision on purpose — it shares the
// array and records the length (#629, clone.go) — which is exactly
// why this one is written down rather than assumed.
func TestCloneDoesNotShareTheCastOrderBackingArray(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]

	castInstantFor(t, g, me)
	castInstantFor(t, g, me)
	castInstantFor(t, g, me)
	early := g.Clone()

	fourth := castInstantFor(t, g, me)
	later := g.Clone()
	if n := len(later.TurnTally.Casts); n != 4 {
		t.Fatalf("the later restore point holds %d casts, want 4", n)
	}

	// Undo back past the fourth cast and play a DIFFERENT spell.
	// RestoreFrom swaps Game.Seats wholesale, so the live seat has to
	// be read again on the far side of it.
	g.WithWriteLock(func() { g.RestoreFrom(early) })
	fifth := castInstantFor(t, g, g.Seats[0])

	if later.TurnTally.Casts[3] == fifth {
		t.Error("the replay wrote into the later restore point's array")
	}
	if later.TurnTally.Casts[3] != fourth {
		t.Errorf("the later restore point's fourth cast changed: got %v, want %v",
			later.TurnTally.Casts[3], fourth)
	}
	if n := len(early.TurnTally.Casts); n != 3 {
		t.Errorf("the earlier restore point holds %d casts, want 3", n)
	}
}

// TestTheCastOrderSurvivesASnapshotRoundTrip — TurnTally rides the
// snapshot as one value, so the field only had to be added to
// cloneTurnTally. This is the assertion that says so.
func TestTheCastOrderSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	castInstantFor(t, g, me)
	storm := castInstantFor(t, g, me)

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	restored.mu.Lock()
	defer restored.mu.Unlock()
	if got := restored.SpellsCastBeforeThisTurn(storm); got != 1 {
		t.Errorf("storm count after a snapshot round trip = %d, want 1", got)
	}
}
