package effects

import (
	"testing"
)

// planes_b11_test.go — Planechase plane cards, batches 11 and 12 (the gold-
// bordered punk and pssc planes, ADR 0216 slice 4, #4870). Every test plays
// the real catalog entry face up on a Planechase table.

var (
	planeSquirrels = planeFixture{"7aac9628-f112-4840-91b4-7bb5cd896685", "We Hope You Like Squirrels", "Plane — Secret Lair"}
)

func TestWeHopeYouLikeSquirrelsGivesTheActingPlayerASquirrel(t *testing.T) {
	requirePlaneRegisteredFull(t, planeSquirrels)
	g, _ := newPlaneGame(t, planeSquirrels)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Test Ritual", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Squirrel" && c.Controller == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d Squirrels for the caster after a spell, want 1", n)
	}
	next := toNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	m := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Squirrel" && c.Controller == next.ID {
			m++
		}
	}
	if m != 1 {
		t.Errorf("the upkeep player has %d Squirrels, want 1", m)
	}
}

func TestWeHopeYouLikeSquirrelsChaosGrowsSquirrels(t *testing.T) {
	g, _ := newPlaneGame(t, planeSquirrels)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sq := b12Creature(g, me.ID, "Squirrel", "Creature — Squirrel", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rollBlank(t, g, me.ID)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if p := tmtB1Card(t, g, sq).CurrentPower(); p != 2 {
		t.Errorf("the Squirrel is %d power, want 2", p)
	}
	if p := tmtB1Card(t, g, bear).CurrentPower(); p != 2 {
		t.Errorf("the Bear is %d power, want 2", p)
	}
}

var planeFblthp = planeFixture{"ec2f6e72-8a2b-44f0-8241-7fe29c601d96", "Fblthp: Completely, Utterly, Totally Lost", "Plane — Secret Lair"}

func TestFblthpMakesAMapOnPlaneswalkAndUpkeep(t *testing.T) {
	requirePlaneRegisteredFull(t, planeFblthp)
	g, _ := newPlaneGame(t, planeFblthp)
	advanceToMain(t, g)
	planeswalkToPlane(t, g, planeFblthp.Oracle)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Map"); n != 1 {
		t.Fatalf("%d Maps after planeswalking to the plane, want 1", n)
	}
	toNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Map"); n != 2 {
		t.Errorf("%d Maps after the next upkeep, want 2", n)
	}
}

func TestFblthpChaosLootsTwo(t *testing.T) {
	g, plane := newPlaneGame(t, planeFblthp)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatal("a blank triggered the chaos ability")
	}
	rollChaos(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 1 {
		t.Fatalf("chaos put %d items on the stack, want 1", n)
	}
}
