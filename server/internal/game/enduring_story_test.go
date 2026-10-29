package game

import (
	"testing"

	"github.com/google/uuid"
)

// enduring_story_test.go — storied and the enduring story (CR 702.195),
// #4475, ADR 0096's 2026-10-29 amendment. The sweep is the city's
// blessing's (earned_designations.go), so these tests are about what is
// different: what counts, and how many.
//
// What a bug here would hide, worst first:
//
//  1. Losing it. The enduring story is kept for the rest of the game
//     (CR 702.195a); nothing may clear it.
//  2. Counting the wrong permanents. Artifacts, Sagas and legendary
//     permanents count, each ONCE however many of the three it is; a
//     plain creature, land or enchantment does not.
//  3. Getting it without storied. CR 702.195a is a static ability of a
//     storied permanent.
//  4. Crossing the two designations: a storied board must not bless, and
//     an ascend board must not give a story.

// seedTyped puts one permanent with the given type line on the
// battlefield under p.
func seedTyped(g *Game, p *Player, typeLine string) uuid.UUID {
	c := NewCard("Test "+typeLine, p.ID)
	c.TypeLine = typeLine
	if c.IsCreature() {
		c.Power, c.Toughness = 1, 1
	}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// seedStoried puts a nonlegendary creature with printed storied on the
// battlefield, so it is not one of the three it counts.
func seedStoried(g *Game, p *Player) uuid.UUID {
	c := NewCard("Test Storyteller", p.ID)
	c.TypeLine = "Creature — Dwarf"
	c.Power, c.Toughness = 1, 1
	c.Keywords = []string{KeywordStoried}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// CR 702.195a: three permanents that are artifacts, Sagas and/or
// legendary, one of each kind here, and not before the third.
func TestStoriedGivesAnEnduringStoryAtThreeArtifactsSagasOrLegendaries(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedStoried(g, me)
	seedTyped(g, me, "Artifact")
	seedTyped(g, me, "Enchantment — Saga")
	settleChecks(g)
	if me.EnduringStory {
		t.Fatal("two counting permanents gave an enduring story")
	}
	seedTyped(g, me, "Legendary Creature — Dwarf")
	settleChecks(g)
	if !me.EnduringStory {
		t.Fatal("an artifact, a Saga and a legendary creature with a storied permanent gave no enduring story")
	}
	if me.CitysBlessing {
		t.Error("storied gave the city's blessing")
	}
}

// Each permanent counts once: a legendary artifact Saga is one, not
// three. Plain creatures, lands and enchantments do not count at all.
func TestEachStoriedPermanentCountsOnce(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedStoried(g, me)
	seedTyped(g, me, "Legendary Artifact — Saga")
	seedTyped(g, me, "Creature — Human")
	seedTyped(g, me, "Land")
	seedTyped(g, me, "Enchantment")
	seedTyped(g, me, "Creature — Dwarf")
	settleChecks(g)
	if me.EnduringStory {
		t.Fatal("one legendary artifact Saga and four non-counting permanents gave an enduring story")
	}
	seedTyped(g, me, "Artifact — Equipment")
	seedTyped(g, me, "Legendary Enchantment")
	settleChecks(g)
	if !me.EnduringStory {
		t.Fatal("three counting permanents gave no enduring story")
	}
}

// The storied permanent counts itself when it is one of the three.
func TestALegendaryStoriedPermanentCountsItself(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedStoried(g, me)
	g.mu.Lock()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].TypeLine = "Legendary Creature — Dwarf"
		}
	}
	g.mu.Unlock()
	seedTyped(g, me, "Artifact")
	seedTyped(g, me, "Artifact")
	settleChecks(g)
	if !me.EnduringStory {
		t.Fatal("a legendary storied creature and two artifacts gave no enduring story")
	}
}

// CR 702.195a: "for the rest of the game". The storied permanent leaving
// and the count falling do not take it away.
func TestTheEnduringStoryIsKeptWhenTheBoardShrinks(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	ids := []uuid.UUID{seedStoried(g, me),
		seedTyped(g, me, "Artifact"), seedTyped(g, me, "Artifact"), seedTyped(g, me, "Artifact")}
	settleChecks(g)
	if !me.EnduringStory {
		t.Fatal("setup: no enduring story")
	}
	g.mu.Lock()
	for _, id := range ids {
		g.Battlefield.Remove(id)
	}
	g.mu.Unlock()
	settleChecks(g)
	if !me.EnduringStory {
		t.Error("the enduring story was lost when the board shrank")
	}
}

// No storied permanent, no story; and the grant is to the storied
// permanent's controller.
func TestTheEnduringStoryNeedsAStoriedPermanentAndGoesToItsController(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	seedStoried(g, me)
	seedTyped(g, me, "Artifact")
	for range 5 {
		seedTyped(g, them, "Legendary Artifact")
	}
	settleChecks(g)
	if me.EnduringStory || them.EnduringStory {
		t.Errorf("enduring story: storied controller with one artifact %v, five legendaries and no storied %v; want false, false",
			me.EnduringStory, them.EnduringStory)
	}
}

// Ten permanents and an ascend permanent bless, but tell no story, even
// when three of the ten are artifacts: the two rows are separate.
func TestAscendDoesNotGiveAnEnduringStory(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedAscender(g, me)
	seedPermanents(g, me, 9) // artifacts
	settleChecks(g)
	if !me.CitysBlessing {
		t.Fatal("setup: no blessing")
	}
	if me.EnduringStory {
		t.Error("ascend gave an enduring story")
	}
}

// An eliminated player is given nothing (CR 800.4a), and the grant is
// announced once (the log line and the layer bump ride the event).
func TestTheEnduringStoryIsAnnouncedOnceAndNotToTheEliminated(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	seedStoried(g, me)
	seedStoried(g, them)
	for range 3 {
		seedTyped(g, me, "Artifact")
		seedTyped(g, them, "Artifact")
	}
	them.Eliminated = true
	settleChecks(g)
	settleChecks(g)
	if them.EnduringStory {
		t.Error("an eliminated player got an enduring story")
	}
	g.mu.Lock()
	again := g.grantEnduringStoryLocked(me, uuid.Nil)
	g.mu.Unlock()
	if again {
		t.Error("granting a second time reported a change")
	}
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventEnduringStory {
			n++
			if ev.Actor != me.ID {
				t.Errorf("EventEnduringStory for %v, want %v", ev.Actor, me.ID)
			}
		}
	}
	if n != 1 {
		t.Errorf("EventEnduringStory emitted %d times, want 1", n)
	}
	g.mu.Lock()
	has, other := g.EnduringStoryForEffect(me.ID), g.EnduringStoryForEffect(uuid.New())
	g.mu.Unlock()
	if !has || other {
		t.Errorf("EnduringStoryForEffect: storied %v, unknown id %v; want true, false", has, other)
	}
}

// The designation rides the snapshot and comes back from a restore.
func TestTheEnduringStorySurvivesASnapshotRestore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	seedStoried(g, me)
	for range 3 {
		seedTyped(g, me, "Artifact")
	}
	settleChecks(g)
	if !me.EnduringStory {
		t.Fatal("setup: no enduring story")
	}
	_, restored := roundTrip(t, g)
	if !restored.Seats[0].EnduringStory {
		t.Error("the enduring story was lost across a snapshot restore")
	}
	if restored.Seats[1].EnduringStory {
		t.Error("a restore gave an enduring story to a seat that had none")
	}
}
