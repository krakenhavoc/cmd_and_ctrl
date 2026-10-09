package game

import (
	"testing"

	"github.com/google/uuid"
)

// citys_blessing_test.go — ascend and the city's blessing (CR 702.131),
// #2696, ADR 0096's 2026-10-08 amendment.
//
// What a bug here would hide, worst first:
//
//  1. Losing it. The blessing is kept for the rest of the game: the
//     catalog used to approximate it with a live "you control ten
//     permanents" read, which shut every card that asked as soon as the
//     board shrank. Nothing may clear it.
//  2. Getting it without ascend. CR 702.131b is a static ability OF AN
//     ASCEND PERMANENT: ten permanents with none is nothing.
//  3. Giving it to the wrong player. The grant is to the ascend
//     permanent's controller.
//  4. The spell half. An instant or sorcery with ascend checks as it
//     resolves (CR 702.131a), before its other instructions.

// seedPermanents puts n vanilla permanents on the battlefield under p.
func seedPermanents(g *Game, p *Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		c := NewCard("Test Permanent", p.ID)
		c.TypeLine = "Artifact"
		g.Battlefield.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

// seedAscender puts a permanent with printed ascend on the battlefield.
func seedAscender(g *Game, p *Player) uuid.UUID {
	c := NewCard("Test Ascender", p.ID)
	c.TypeLine = "Creature — Test"
	c.Power, c.Toughness = 1, 1
	c.Keywords = []string{KeywordAscend}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func settleChecks(g *Game) {
	g.mu.Lock()
	g.stateBasedActionsLocked()
	g.mu.Unlock()
}

// CR 702.131b: the ascend permanent's controller gets the blessing as
// soon as they control ten permanents, and not before.
func TestAscendPermanentGrantsTheBlessingAtTen(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedAscender(g, me)
	seedPermanents(g, me, 8)
	settleChecks(g)
	if me.CitysBlessing {
		t.Fatal("nine permanents gave the city's blessing")
	}
	seedPermanents(g, me, 1)
	settleChecks(g)
	if !me.CitysBlessing {
		t.Fatal("ten permanents with an ascend permanent did not give the city's blessing")
	}
}

// CR 702.131c: it is kept when the ascend permanent leaves and the
// board shrinks. The state-based pass that follows must not clear it.
func TestTheBlessingIsKeptWhenTheBoardShrinks(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	asc := seedAscender(g, me)
	ids := seedPermanents(g, me, 9)
	settleChecks(g)
	if !me.CitysBlessing {
		t.Fatal("setup: no blessing at ten permanents")
	}
	g.mu.Lock()
	g.Battlefield.Remove(asc)
	for _, id := range ids {
		g.Battlefield.Remove(id)
	}
	g.mu.Unlock()
	settleChecks(g)
	if !me.CitysBlessing {
		t.Error("the city's blessing was lost when the board shrank")
	}
	if g.permanentsControlledLocked(me.ID) != 0 {
		t.Fatal("setup: the board did not empty")
	}
}

// CR 702.131b is a static ability OF an ascend permanent: ten
// permanents and no ascend is nothing.
func TestTenPermanentsWithoutAscendIsNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedPermanents(g, me, 12)
	settleChecks(g)
	if me.CitysBlessing {
		t.Error("ten permanents with no ascend gave the city's blessing")
	}
}

// The grant is to the ascend permanent's controller, not to whoever
// else has a wide board or whoever the permanent's owner is.
func TestTheBlessingGoesToTheAscendPermanentsController(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	seedAscender(g, me)
	seedPermanents(g, me, 3)
	seedPermanents(g, them, 12)
	settleChecks(g)
	if me.CitysBlessing {
		t.Error("a controller with four permanents got the blessing")
	}
	if them.CitysBlessing {
		t.Error("a player with twelve permanents and no ascend got the blessing")
	}
}

// A permanent changing hands moves the static with it: its new
// controller is the one who needs ten permanents.
func TestAscendFollowsTheControllerOfThePermanent(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	asc := seedAscender(g, me)
	seedPermanents(g, them, 9)
	g.mu.Lock()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == asc {
			g.Battlefield.Cards[i].Controller = them.ID
		}
	}
	g.mu.Unlock()
	settleChecks(g)
	if !them.CitysBlessing || me.CitysBlessing {
		t.Errorf("blessing after a control change: new controller %v, old controller %v; want true, false",
			them.CitysBlessing, me.CitysBlessing)
	}
}

// A player who has left the game is not given anything (CR 800.4a).
func TestAnEliminatedPlayerIsNotGivenTheBlessing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedAscender(g, me)
	seedPermanents(g, me, 9)
	me.Eliminated = true
	settleChecks(g)
	if me.CitysBlessing {
		t.Error("an eliminated player got the city's blessing")
	}
}

// The grant announces itself once, and only on a real change.
func TestTheBlessingIsAnnouncedOnce(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedAscender(g, me)
	seedPermanents(g, me, 9)
	count := func() int {
		n := 0
		for _, ev := range g.Events {
			if ev.Kind == EventCitysBlessing && ev.Actor == me.ID {
				n++
			}
		}
		return n
	}
	settleChecks(g)
	settleChecks(g)
	g.mu.Lock()
	again := g.grantCitysBlessingLocked(me, uuid.Nil)
	g.mu.Unlock()
	if again {
		t.Error("granting a second time reported a change")
	}
	if got := count(); got != 1 {
		t.Errorf("EventCitysBlessing emitted %d times, want 1", got)
	}
}

// seedAscendSpell puts an ascend sorcery in the seat's hand.
func seedAscendSpell(me *Player) Card {
	c := NewCard("Test Ascend Sorcery", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{0}"
	c.Keywords = []string{KeywordAscend}
	me.Hand.PushTop(c)
	return c
}

// CR 702.131a: an instant or sorcery with ascend checks as it
// resolves.
func TestAscendSpellChecksAtResolution(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	seedPermanents(g, me, 10)
	c := seedAscendSpell(me)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if me.CitysBlessing {
		t.Fatal("the blessing arrived on cast, before the spell resolved")
	}
	resolveTop(t, g)
	if !me.CitysBlessing {
		t.Error("a resolving ascend sorcery with ten permanents did not give the blessing")
	}
}

func TestAscendSpellWithNineDoesNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	seedPermanents(g, me, 9)
	c := seedAscendSpell(me)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	if me.CitysBlessing {
		t.Error("an ascend sorcery with nine permanents gave the blessing")
	}
}

// A spell without ascend does not check, however wide the board. (The
// spell is not an ascend permanent either, so the sweep gives nothing.)
func TestASpellWithoutAscendDoesNotCheck(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	seedPermanents(g, me, 12)
	c := seedAscendSpell(me)
	me.Hand.Cards[len(me.Hand.Cards)-1].Keywords = nil
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	if me.CitysBlessing {
		t.Error("a spell without ascend gave the blessing")
	}
}

// The designation rides the snapshot and the clone behind undo.
func TestTheBlessingSurvivesSnapshotAndClone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.CitysBlessing = true

	_, restored := roundTrip(t, g)
	if !restored.Seats[0].CitysBlessing {
		t.Error("the blessing was lost in a snapshot round trip")
	}
	if restored.Seats[1].CitysBlessing {
		t.Error("the snapshot gave the blessing to the wrong seat")
	}
	if !g.Clone().Seats[0].CitysBlessing {
		t.Error("the blessing was lost in a clone")
	}
}

// The reader card code uses.
func TestCitysBlessingForEffectNamesNoOneForAnUnknownID(t *testing.T) {
	g := newActiveGame(t)
	g.Seats[0].CitysBlessing = true
	if !g.CitysBlessingForEffect(g.Seats[0].ID) {
		t.Error("the blessed seat reads as unblessed")
	}
	if g.CitysBlessingForEffect(g.Seats[1].ID) || g.CitysBlessingForEffect(uuid.New()) || g.CitysBlessingForEffect(uuid.Nil) {
		t.Error("an unblessed seat or an unknown ID reads as blessed")
	}
}
