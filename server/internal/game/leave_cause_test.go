package game

import (
	"testing"

	"github.com/google/uuid"
)

// leave_cause_test.go — #4395, ADR 0209 §3: every way a permanent leaves
// the battlefield names why on its EventZoneMove, set at the two movers
// and not per card.

// leaveCreature puts a 2/2 on the battlefield for `owner`.
func leaveCreature(g *Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	return id
}

// leaveEvents returns the EventZoneMove events that took one of ids off
// the battlefield, in emission order.
func leaveEvents(g *Game, ids ...uuid.UUID) []Event {
	want := map[uuid.UUID]bool{}
	for _, id := range ids {
		want[id] = true
	}
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventZoneMove && ev.OldZone == ZoneBattlefield && want[ev.CardID] {
			out = append(out, ev)
		}
	}
	return out
}

func wantLeave(t *testing.T, g *Game, id uuid.UUID, cause LeaveCause) Event {
	t.Helper()
	evs := leaveEvents(g, id)
	if len(evs) != 1 {
		t.Fatalf("got %d battlefield zone moves for the card, want 1", len(evs))
	}
	if evs[0].LeaveCause != cause {
		t.Errorf("LeaveCause = %q, want %q (to %s)", evs[0].LeaveCause, cause, evs[0].NewZone)
	}
	return evs[0]
}

func TestLeaveCauseByWayOffTheBattlefield(t *testing.T) {
	t.Run("destroy by an effect", func(t *testing.T) {
		g := newActiveGame(t)
		id := leaveCreature(g, g.Seats[0].ID, "Victim")
		g.WithWriteLock(func() { g.destroyPermanentsLocked([]uuid.UUID{id}, DestroyOptions{}) })
		wantLeave(t, g, id, LeaveCauseDestroy)
	})
	t.Run("lethal damage is a state-based action", func(t *testing.T) {
		g := newActiveGame(t)
		id := leaveCreature(g, g.Seats[0].ID, "Victim")
		if err := g.MarkDamage(id, 2); err != nil {
			t.Fatal(err)
		}
		// CR 704.5g destroys it, and the cause is still the SBA: that
		// is the "dies" look, not the "destroyed by a spell" one.
		wantLeave(t, g, id, LeaveCauseSBA)
	})
	t.Run("zero toughness is a state-based action", func(t *testing.T) {
		g := newActiveGame(t)
		id := leaveCreature(g, g.Seats[0].ID, "Victim")
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Toughness = 0
					// A real printing is what tells CR 704.5f that a 0
					// toughness is a real 0 (#691).
					g.Battlefield.Cards[i].ScryfallID = "test-printing"
				}
			}
			g.runStateChecksLocked()
		})
		wantLeave(t, g, id, LeaveCauseSBA)
	})
	t.Run("the legend rule is a state-based action", func(t *testing.T) {
		g := newFourPlayerActiveGame(t)
		me := g.Seats[0].ID
		keep := pushLegendForTest(g, me, "Teferi, Temporal Pilgrim", "Legendary Planeswalker — Teferi")
		drop := pushLegendForTest(g, me, "Teferi, Temporal Pilgrim", "Legendary Planeswalker — Teferi")
		g.WithWriteLock(func() { g.runStateChecksLocked() })
		ch := legendPromptFor(g, me)
		if ch == nil {
			t.Fatal("no legend-rule prompt was queued")
		}
		if err := g.ResolveLegendRule(ch.ID, me, keep); err != nil {
			t.Fatal(err)
		}
		wantLeave(t, g, drop, LeaveCauseSBA)
	})
	t.Run("sacrifice", func(t *testing.T) {
		g := newActiveGame(t)
		id := leaveCreature(g, g.Seats[0].ID, "Victim")
		if err := g.SacrificePermanent(g.Seats[0].ID, id); err != nil {
			t.Fatal(err)
		}
		wantLeave(t, g, id, LeaveCauseSacrifice)
	})
	t.Run("exile by an effect", func(t *testing.T) {
		g := newActiveGame(t)
		id := leaveCreature(g, g.Seats[0].ID, "Victim")
		g.WithWriteLock(func() { g.ExileCardsForEffect([]uuid.UUID{id}) })
		ev := wantLeave(t, g, id, LeaveCauseExile)
		if ev.NewZone != ZoneExile {
			t.Errorf("NewZone = %s, want exile", ev.NewZone)
		}
	})
	t.Run("bounce to hand", func(t *testing.T) {
		g := newActiveGame(t)
		id := leaveCreature(g, g.Seats[0].ID, "Victim")
		g.WithWriteLock(func() { g.BounceCardsToHandForEffect([]uuid.UUID{id}) })
		ev := wantLeave(t, g, id, LeaveCauseBounce)
		if ev.NewZone != ZoneHand {
			t.Errorf("NewZone = %s, want hand", ev.NewZone)
		}
	})
	t.Run("put into a library is other", func(t *testing.T) {
		g := newActiveGame(t)
		id := leaveCreature(g, g.Seats[0].ID, "Victim")
		g.WithWriteLock(func() {
			if err := g.TuckToLibraryForEffect(id, true); err != nil {
				t.Fatal(err)
			}
		})
		ev := wantLeave(t, g, id, LeaveCauseOther)
		if ev.NewZone != ZoneLibrary {
			t.Errorf("NewZone = %s, want library", ev.NewZone)
		}
	})
	t.Run("a sandbox move is other", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		id := leaveCreature(g, me.ID, "Victim")
		err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, id)
		if err != nil {
			t.Fatal(err)
		}
		wantLeave(t, g, id, LeaveCauseOther)
	})
	t.Run("a card that is not a battlefield exit says nothing", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		c := NewCard("In hand", me.ID)
		me.Hand.PushTop(c)
		err := g.MoveCardByID(ZoneRef{Kind: ZoneHand, Owner: me.ID}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, c.InstanceID)
		if err != nil {
			t.Fatal(err)
		}
		for _, ev := range g.Events {
			if ev.Kind == EventZoneMove && ev.CardID == c.InstanceID && ev.LeaveCause != "" {
				t.Errorf("a hand-to-graveyard move carries LeaveCause %q", ev.LeaveCause)
			}
		}
	})
}

// A wipe's deaths share a batch and a source: the source is the card
// whose resolution the batch is, read from the resolving slot, and it is
// still there for the state-based action that follows the resolution.
func TestLeaveSourceIsTheResolvingCardForEffectsAndForTheSBAThatFollows(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	a, b := leaveCreature(g, me, "A"), leaveCreature(g, opp, "B")
	spell := uuid.New()
	g.WithWriteLock(func() {
		g.resolving = &resolvingItem{item: &StackItem{ID: spell, SourceCardID: spell}, sourceEpoch: -1}
		g.destroyPermanentsLocked([]uuid.UUID{a}, DestroyOptions{})
	})
	if err := g.MarkDamage(b, 2); err != nil {
		t.Fatal(err)
	}
	evs := leaveEvents(g, a, b)
	if len(evs) != 2 {
		t.Fatalf("got %d exits, want 2", len(evs))
	}
	for _, ev := range evs {
		if ev.LeaveSource != spell {
			t.Errorf("%s: LeaveSource = %s, want the resolving spell %s", ev.LeaveCause, ev.LeaveSource, spell)
		}
	}
	if evs[0].Batch != evs[1].Batch {
		t.Errorf("batches %d and %d differ for one resolution", evs[0].Batch, evs[1].Batch)
	}
	if evs[0].LeaveCause != LeaveCauseDestroy || evs[1].LeaveCause != LeaveCauseSBA {
		t.Errorf("causes = %q, %q; want destroy, sba", evs[0].LeaveCause, evs[1].LeaveCause)
	}

	// Play moves on: the next batch has no resolving item, so a state-based
	// action in it (combat damage) has no single source.
	c := leaveCreature(g, me, "C")
	g.BeginEventBatchForTest()
	if err := g.MarkDamage(c, 2); err != nil {
		t.Fatal(err)
	}
	if ev := wantLeave(t, g, c, LeaveCauseSBA); ev.LeaveSource != uuid.Nil {
		t.Errorf("a death with nothing resolving has LeaveSource %s, want none", ev.LeaveSource)
	}
}

// A sacrifice paid as a cost is nobody's resolution, so it names no
// source, while one an effect asks for does.
func TestLeaveSacrificeAnnouncementNamesTheResolvingCardUnlessItIsACost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	spell := uuid.New()
	effect, cost := leaveCreature(g, me, "Effect"), leaveCreature(g, me, "Cost")
	g.WithWriteLock(func() {
		g.resolving = &resolvingItem{item: &StackItem{ID: spell, SourceCardID: spell}, sourceEpoch: -1}
		if err := g.sacrificePermanentLocked(effect); err != nil {
			t.Fatal(err)
		}
		if err := g.sacrificeAnsweredLocked(cost, commanderZoneUnasked); err != nil {
			t.Fatal(err)
		}
	})
	sources := map[uuid.UUID]uuid.UUID{}
	for _, ev := range g.Events {
		if ev.Kind == EventSacrifice {
			sources[ev.CardID] = ev.LeaveSource
		}
	}
	if sources[effect] != spell {
		t.Errorf("an effect's sacrifice: LeaveSource = %s, want %s", sources[effect], spell)
	}
	if sources[cost] != uuid.Nil {
		t.Errorf("a cost's sacrifice: LeaveSource = %s, want none", sources[cost])
	}
	for _, ev := range leaveEvents(g, effect, cost) {
		want := spell
		if ev.CardID == cost {
			want = uuid.Nil
		}
		if ev.LeaveCause != LeaveCauseSacrifice || ev.LeaveSource != want {
			t.Errorf("sacrificed card's zone move = (%q, %s), want (sacrifice, %s)", ev.LeaveCause, ev.LeaveSource, want)
		}
	}
}

// A destruction a replacement turns into an exile is still the verb that
// was asked for; where it went is NewZone, which a client reads first.
func TestLeaveCauseIsTheInstructionNotTheDestination(t *testing.T) {
	g := newActiveGame(t)
	id := wipeTarget(g, g.Seats[0])
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(leaveBattlefieldReplacement(id,
			"exile it instead", func(ev *ReplacementEvent) {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
			}))
		g.destroyPermanentsLocked([]uuid.UUID{id}, DestroyOptions{})
	})
	ev := wantLeave(t, g, id, LeaveCauseDestroy)
	if ev.NewZone != ZoneExile {
		t.Errorf("NewZone = %s, want exile", ev.NewZone)
	}
}
