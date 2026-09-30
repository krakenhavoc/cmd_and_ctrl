package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// discover_test.go — CR 701.57 and ADR 0099. What a bug here would
// hide, worst first:
//
//  1. The two rules discover differs from cascade on: a card with mana
//     value EQUAL to N is a hit, and an uncast hit goes to the HAND.
//  2. The window. A free cast held past the discoverer's next pass, and
//     still ending in hand, is strictly stronger than printed (owner
//     decision 1).
//  3. The cap. "If the resulting spell's mana value is N or less" is
//     judged against the face actually cast.
//  4. When "whenever you discover" fires: at the settle, and on a cast
//     after the EventCast (owner decision 3).

func discoverEvents(g *Game) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventDiscover {
			out = append(out, ev)
		}
	}
	return out
}

func discoverPrompt(t *testing.T, g *Game) *PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMayCast {
			return c
		}
	}
	t.Fatalf("no may_cast prompt outstanding")
	return nil
}

// The stopping rule: lands and cards over N are passed over, and a card
// whose mana value is exactly N is a hit (CR 701.57a — "N or less",
// where cascade's is "less than").
func TestDiscoverStopsOnTheFirstNonlandWithManaValueNOrLess(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	deep := libCard(me.ID, "Never Reached", "Instant", "{U}")
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	tooBig := libCard(me.ID, "Big Thing", "Creature — Beast", "{3}{G}")
	land := libCard(me.ID, "Island", "Basic Land — Island", "")
	libraryOf(t, me, deep, hit, tooBig, land)

	source := uuid.New()
	g.WithWriteLock(func() {
		if err := g.DiscoverThenForEffect(me.ID, source, 3, nil); err != nil {
			t.Fatalf("DiscoverThenForEffect: %v", err)
		}
	})
	c := discoverPrompt(t, g)
	if c.MayCastCard != hit.InstanceID {
		t.Fatalf("discover hit the wrong card")
	}
	if c.MayCastKeyword != MayCastKeywordDiscover || c.AcceptLabel == "" || c.DeclineLabel == "" {
		t.Errorf("prompt words = %q / %q / %q, want discover's", c.MayCastKeyword, c.AcceptLabel, c.DeclineLabel)
	}
	if !me.Library.Contains(deep.InstanceID) {
		t.Errorf("discover exiled past its hit")
	}
	if got := discoverEvents(g); len(got) != 0 {
		t.Errorf("EventDiscover fired before the discover was complete: %v", got)
	}
}

// Declining: the hit goes into the discoverer's HAND (not the bottom),
// the rest go to the bottom, the discover is announced with the card
// and N, and the continuation sees the discovered card.
func TestDiscoverDeclinePutsTheCardIntoHand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	floor := libCard(me.ID, "Bottom Card", "Instant", "{U}")
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	land := libCard(me.ID, "Island", "Basic Land — Island", "")
	libraryOf(t, me, floor, hit, land)

	source := uuid.New()
	var got DiscoverResult
	ran := 0
	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(me.ID, source, 4, func(_ *Game, r DiscoverResult) error {
			got = r
			ran++
			return nil
		})
	})
	if ran != 0 {
		t.Fatalf("Then ran before the prompt was answered")
	}
	answerMayCast(t, g, me.ID, false)

	if !me.Hand.Contains(hit.InstanceID) {
		t.Fatalf("declined discover hit is not in hand")
	}
	ids := libraryIDs(me)
	if len(ids) != 2 || ids[0] != land.InstanceID || ids[1] != floor.InstanceID {
		t.Errorf("library (bottom first) = %v, want [Island, Bottom Card]", ids)
	}
	if ran != 1 || got.Discovered != hit.InstanceID || got.N != 4 || got.ManaValue != 3 {
		t.Errorf("Then ran %d times with %+v", ran, got)
	}
	evs := discoverEvents(g)
	if len(evs) != 1 || evs[0].Actor != me.ID || evs[0].Source != source || evs[0].Amount != 4 || evs[0].CardID != hit.InstanceID {
		t.Errorf("EventDiscover = %+v, want one naming the hit, N and source", evs)
	}
}

// Accepting: a {0}, flash-timed, capped grant (owner decisions 1 and 2's
// shared pieces); casting it completes the discover right after the
// EventCast; the grant is spent, so the next pass does nothing to it.
func TestDiscoverAcceptGrantsACappedFlashFreeCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	// A sorcery, discovered during the upkeep: CR 608.2g lets it be
	// cast then, which the grant's TimingFlash is.
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	libraryOf(t, me, hit)

	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(me.ID, uuid.New(), 3, nil)
	})
	answerMayCast(t, g, me.ID, true)

	perm := g.CastPermissionOnCardByIDForEffect(hit.InstanceID)
	if perm == nil || !g.CastPermissionActiveForEffect(perm, me.ID) {
		t.Fatalf("no live permission on the discovered card")
	}
	if perm.Cost != "{0}" || perm.Timing != TimingFlash || !perm.CastOnly {
		t.Errorf("grant = cost %q timing %q castOnly %v, want {0} / flash / cast-only", perm.Cost, perm.Timing, perm.CastOnly)
	}
	if perm.MaxSpellManaValue == nil || *perm.MaxSpellManaValue != 3 || perm.LapseOnPass != LapseToHand || perm.Discover == nil {
		t.Errorf("grant cap/lapse/discover = %v / %q / %v", perm.MaxSpellManaValue, perm.LapseOnPass, perm.Discover)
	}
	if len(discoverEvents(g)) != 0 {
		t.Fatalf("EventDiscover fired at the answer, before the card was settled")
	}

	if g.Turn.Step == StepPrecombatMain || g.Turn.Step == StepPostcombatMain {
		t.Fatalf("test needs a non-main step to prove the flash timing")
	}
	if err := g.CastSpell(me.ID, hit.InstanceID, CastSpellParams{FromZone: string(ZoneExile)}); err != nil {
		t.Fatalf("casting the discovered sorcery outside a main phase: %v", err)
	}
	// EventDiscover comes after the EventCast of the discovered spell.
	castAt, discoverAt := -1, -1
	for i, ev := range g.Events {
		if ev.Kind == EventCast && ev.CardID == hit.InstanceID {
			castAt = i
		}
		if ev.Kind == EventDiscover {
			discoverAt = i
		}
	}
	if castAt < 0 || discoverAt < 0 || discoverAt < castAt {
		t.Errorf("EventCast at %d, EventDiscover at %d: want the discover after the cast", castAt, discoverAt)
	}
	for _, perm := range me.CastPermissions {
		if perm.LapseOnPass != "" {
			t.Errorf("the spent grant is still held: %+v", perm)
		}
	}
	// A pass now closes nothing and announces nothing more.
	g.WithWriteLock(func() {
		if g.closePassWindowsLocked(me.ID) {
			t.Errorf("a spent grant lapsed again on the pass")
		}
	})
	if n := len(discoverEvents(g)); n != 1 {
		t.Errorf("EventDiscover fired %d times, want 1", n)
	}
}

// Owner decision 1: accepting and then passing without casting is the
// decline, run late — the card goes to hand and the discover completes.
func TestDiscoverGrantLapsesIntoHandOnTheNextPass(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	libraryOf(t, me, hit)
	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(me.ID, uuid.New(), 3, nil)
	})
	answerMayCast(t, g, me.ID, true)
	if g.Seats[g.Turn.PriorityHolder].ID != me.ID {
		t.Fatalf("test needs the discoverer to hold priority")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if !me.Hand.Contains(hit.InstanceID) || g.Exile.Contains(hit.InstanceID) {
		t.Fatalf("an uncast discovered card did not go to hand on the pass")
	}
	if g.CastPermissionOnCardByIDForEffect(hit.InstanceID) != nil {
		t.Errorf("the grant outlived the pass")
	}
	evs := discoverEvents(g)
	if len(evs) != 1 || evs[0].CardID != hit.InstanceID || evs[0].Amount != 3 {
		t.Errorf("EventDiscover = %+v, want one on the lapse", evs)
	}
}

// Another seat's pass does not close the discoverer's window.
func TestDiscoverGrantIgnoresOtherPlayersPasses(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hit := libCard(opp.ID, "Divination", "Sorcery", "{2}{U}")
	libraryOf(t, opp, hit)
	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(opp.ID, uuid.New(), 3, nil)
	})
	answerMayCast(t, g, opp.ID, true)
	if g.Seats[g.Turn.PriorityHolder].ID != me.ID {
		t.Fatalf("test needs the active player, not the discoverer, to hold priority")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if !g.Exile.Contains(hit.InstanceID) || g.CastPermissionOnCardByIDForEffect(hit.InstanceID) == nil {
		t.Fatalf("the active player's pass closed the discoverer's window")
	}
}

// A "whenever you discover" trigger raised by the lapse keeps priority
// with the discoverer, with the trigger on the stack — as if the card
// had gone to hand during the resolution.
func TestDiscoverLapseThatTriggersKeepsPriority(t *testing.T) {
	const watcher = "test-discover-watcher"
	prev := CatalogTriggers
	CatalogTriggers = func(id string) []TriggeredAbility {
		if id != watcher {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventDiscover},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return newTriggeredItemForTest(source, "whenever you discover", func(*Game, *StackItem) error { return nil })
			},
		}}
	}
	t.Cleanup(func() { CatalogTriggers = prev })

	g := newActiveGame(t)
	me := g.Seats[0]
	w := NewCard("Watcher", me.ID)
	w.TypeLine = "Creature — Human"
	w.OracleID = watcher
	w.Controller = me.ID
	g.Battlefield.PushTop(w)
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	libraryOf(t, me, hit)
	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(me.ID, uuid.New(), 3, nil)
	})
	answerMayCast(t, g, me.ID, true)
	holder := g.Turn.PriorityHolder
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if g.Turn.PriorityHolder != holder {
		t.Errorf("priority moved on although the lapse put a trigger on the stack")
	}
	found := false
	for _, it := range g.StackMeta {
		if it.Kind == StackItemTriggered {
			found = true
		}
	}
	if !found {
		t.Errorf("the whenever-you-discover trigger is not on the stack")
	}
}

// Nothing to discover, and an empty library: everything goes back, the
// player has still discovered (CR 701.57b), and Then runs with no card.
func TestDiscoverWithNoHitStillDiscovers(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	big := libCard(me.ID, "Big Thing", "Creature — Beast", "{5}{G}")
	land := libCard(me.ID, "Forest", "Basic Land — Forest", "")
	libraryOf(t, me, big, land)
	var got *DiscoverResult
	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(me.ID, uuid.New(), 2, func(_ *Game, r DiscoverResult) error {
			got = &r
			return nil
		})
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMayCast {
			t.Fatalf("prompted with no hit")
		}
	}
	if me.Library.Size() != 2 || len(g.Exile.Cards) != 0 {
		t.Errorf("library %d, exile %d: want everything back", me.Library.Size(), len(g.Exile.Cards))
	}
	if got == nil || got.Discovered != uuid.Nil || got.ManaValue != 0 {
		t.Errorf("Then = %+v, want no discovered card", got)
	}
	evs := discoverEvents(g)
	if len(evs) != 1 || evs[0].CardID != uuid.Nil || evs[0].Amount != 2 {
		t.Errorf("EventDiscover = %+v, want one with no card", evs)
	}

	me.Library.Cards = nil
	ran := false
	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(me.ID, uuid.New(), 4, func(*Game, DiscoverResult) error {
			ran = true
			return nil
		})
	})
	if !ran || len(discoverEvents(g)) != 2 {
		t.Errorf("empty library: Then ran %v, events %d", ran, len(discoverEvents(g)))
	}
}

// The cap reads the face actually being cast: a modal DFC discovered
// off its cheap front face cannot be cast as its expensive back face.
func TestDiscoverCapRefusesAnExpensiveBackFace(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Cheap Front", me.ID)
	c.Layout = LayoutModalDFC
	c.Faces = []Face{
		{Name: "Cheap Front", TypeLine: "Creature — Cat", ManaCost: "{1}{R}", Power: 2, Toughness: 1},
		{Name: "Pricey Back", TypeLine: "Creature — Devil", ManaCost: "{5}{R}{R}", Power: 7, Toughness: 7},
	}
	c.SetFace(0)
	libraryOf(t, me, c)
	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(me.ID, uuid.New(), 2, nil)
	})
	answerMayCast(t, g, me.ID, true)

	err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{FromZone: string(ZoneExile), Face: 1})
	if !errors.Is(err, ErrSpellManaValueTooHigh) {
		t.Fatalf("back face cast: err = %v, want ErrSpellManaValueTooHigh", err)
	}
	card, _ := g.LookupCardForEffect(c.InstanceID)
	back := card
	back.SetFace(1)
	g.WithWriteLock(func() {
		perm := g.CastPermissionForLocked(me.ID, back, ZoneExile)
		if offers := g.CastOffersForLocked(me.ID, back, ZoneExile, perm); len(offers) != 0 {
			t.Errorf("the offer list still lists a back-face cast: %v", offers)
		}
	})
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{FromZone: string(ZoneExile), Face: 0}); err != nil {
		t.Fatalf("front face cast: %v", err)
	}
}

// A discoverer who has left the game is not asked; the discover still
// runs its continuation.
func TestDiscoverByAPlayerWhoLeftRunsThen(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Eliminated = true
	ran := false
	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(me.ID, uuid.New(), 3, func(*Game, DiscoverResult) error {
			ran = true
			return nil
		})
	})
	if !ran {
		t.Errorf("Then did not run for a departed discoverer")
	}
}

// The end-of-turn backstop: a grant the sandbox walked past without a
// pass (pass_turn) lapses at the sweep rather than stranding the card.
func TestDiscoverGrantLapsesAtTheEndOfTurnSweep(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	libraryOf(t, me, hit)
	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(me.ID, uuid.New(), 3, nil)
	})
	answerMayCast(t, g, me.ID, true)
	g.WithWriteLock(func() {
		g.sweepCastPermissionsLocked(true)
	})
	if !me.Hand.Contains(hit.InstanceID) {
		t.Errorf("the sweep dropped the grant without putting the card into hand")
	}
	if len(discoverEvents(g)) != 1 {
		t.Errorf("the sweep's lapse did not complete the discover")
	}
}

// ADR 0099 §8: an open discover prompt is a continuation (counted, like
// cascade's), and once it is answered the table is a restore point
// again — the grant is plain data and survives the round trip whole.
func TestDiscoverGrantSurvivesARestore(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0]
	hit := libCard(me.ID, "Divination", "Sorcery", "{2}{U}")
	libraryOf(t, me, hit)
	source := uuid.New()
	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(me.ID, source, 3, nil)
	})
	if g.CaptureSnapshot().Restorable() {
		t.Fatalf("a table with the discover prompt open claims to be a restore point")
	}
	answerMayCast(t, g, me.ID, true)
	snap, restored := roundTrip(t, g)
	if !snap.Restorable() {
		t.Fatalf("an answered discover left the table unrestorable: %+v", snap.Continuations)
	}
	var perm *CastPermission
	for _, p := range restored.Seats {
		if p.ID != me.ID {
			continue
		}
		for i := range p.CastPermissions {
			if p.CastPermissions[i].Discover != nil {
				perm = &p.CastPermissions[i]
			}
		}
	}
	if perm == nil || perm.LapseOnPass != LapseToHand || perm.MaxSpellManaValue == nil || *perm.MaxSpellManaValue != 3 ||
		perm.Discover.N != 3 || perm.Discover.Source != source {
		t.Fatalf("restored discover grant = %+v", perm)
	}
}
