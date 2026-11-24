package game

import (
	"testing"

	"github.com/google/uuid"
)

// haunt_test.go — #6083: the engine half of haunt (CR 702.55), the link
// between an exiled card and the creature it haunts. The card-level tests
// (Blind Hunter, Orzhov Euthanist, Orzhov Pontiff) live in cards/effects.

func seedHauntCreature(g *Game, owner *Player, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Spirit"
	c.Power, c.Toughness = 2, 2
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func hauntExiled(g *Game, id uuid.UUID) *Card {
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == id {
			return &g.Exile.Cards[i]
		}
	}
	return nil
}

func TestHauntLinksTheExiledCardToTheObjectItHaunts(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	me, opp := g.Seats[0], g.Seats[1]
	var haunter, victim, other uuid.UUID
	g.WithWriteLock(func() {
		haunter = seedHauntCreature(g, me, "Haunter")
		victim = seedHauntCreature(g, opp, "Victim")
		other = seedHauntCreature(g, opp, "Other")
	})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, haunter); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if !g.HauntForEffect(haunter, victim) {
			t.Fatal("HauntForEffect reported no haunt")
		}
		c := hauntExiled(g, haunter)
		if c == nil || !g.HauntsForEffect(c) || c.HauntedID != victim {
			t.Fatalf("exiled %+v, want a card haunting the victim", c)
		}
	})
	// Another creature dying is not the creature it haunts.
	from := len(g.Events)
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: opp.ID}, other); err != nil {
		t.Fatal(err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: opp.ID}, victim); err != nil {
		t.Fatal(err)
	}
	var died []uuid.UUID
	g.WithWriteLock(func() {
		c := hauntExiled(g, haunter)
		for _, ev := range g.Events[from:] {
			if g.CreatureItHauntsDiedForEffect(ev, c) {
				died = append(died, ev.CardID)
			}
		}
	})
	if len(died) != 1 || died[0] != victim {
		t.Errorf("the haunted-creature-died condition held for %v, want only the victim", died)
	}
}

func TestHauntEndsWhenEitherObjectEnds(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	me, opp := g.Seats[0], g.Seats[1]
	var haunter, victim uuid.UUID
	g.WithWriteLock(func() {
		haunter = seedHauntCreature(g, me, "Haunter")
		victim = seedHauntCreature(g, opp, "Victim")
	})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, haunter); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.HauntForEffect(haunter, victim) })
	// The haunting card leaves exile and comes back: a new object (CR 400.7).
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneExile}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, haunter); err != nil {
		t.Fatal(err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, ZoneRef{Kind: ZoneExile}, haunter); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if c := hauntExiled(g, haunter); c == nil || g.HauntsForEffect(c) {
			t.Error("a card that left exile and came back still haunts")
		}
	})
}

func TestHauntNeedsTheTargetOnTheBattlefield(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	me := g.Seats[0]
	var haunter uuid.UUID
	g.WithWriteLock(func() { haunter = seedHauntCreature(g, me, "Haunter") })
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, haunter); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if g.HauntForEffect(haunter, uuid.New()) {
			t.Error("haunted a creature that is not on the battlefield")
		}
	})
	if !me.Graveyard.Contains(haunter) {
		t.Error("the card left the graveyard with nothing to haunt")
	}
}

// resolvedMoveOf is the EventZoneMove that took `id` from the stack to a
// graveyard after `from`, and whether there was one.
func resolvedMoveOf(g *Game, from int, id uuid.UUID) (Event, bool) {
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventZoneMove && ev.CardID == id && ev.OldZone == ZoneStack && ev.NewZone == ZoneGraveyard {
			return ev, true
		}
	}
	return Event{}, false
}

// CR 702.55a's second sentence (#6083): haunt on a spell triggers when the
// spell "is put into a graveyard during its resolution". The move a
// resolving instant or sorcery makes says so (Event.Resolved); the same
// move made by CR 608.2b, a spell whose every target is gone, does not.
func TestASpellPutIntoAGraveyardAsItResolvesIsMarkedResolved(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Sorcery", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{B}"
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	from := len(g.Events)
	resolveTop(t, g)
	ev, ok := resolvedMoveOf(g, from, c.InstanceID)
	if !ok {
		t.Fatal("the resolved sorcery made no move from the stack to the graveyard")
	}
	if !ev.Resolved {
		t.Error("the resolving sorcery's move to the graveyard is not marked Resolved")
	}
}

func TestASpellWhoseTargetsAreAllGoneIsNotMarkedResolved(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Instant", me.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{B}"
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Targets: []TargetRef{{Kind: TargetPlayer, ID: them.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	// Its one target leaves the game (CR 800.4a), so the spell is removed
	// as it would resolve (CR 608.2b).
	if err := g.Concede(them.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	from := len(g.Events)
	resolveTop(t, g)
	ev, ok := resolvedMoveOf(g, from, c.InstanceID)
	if !ok {
		t.Fatal("the fizzled instant made no move from the stack to the graveyard")
	}
	if ev.Resolved {
		t.Error("a spell removed for want of legal targets (CR 608.2b) is marked Resolved")
	}
}
