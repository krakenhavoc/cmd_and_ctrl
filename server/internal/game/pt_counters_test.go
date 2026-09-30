package game

import (
	"testing"

	"github.com/google/uuid"
)

// pt_counters_test.go — #1664: every P/T counter kind changes power and
// toughness (CR 122.1a); only +1/+1 and -1/-1 annihilate (CR 704.5q).

func TestParsePTCounter(t *testing.T) {
	cases := []struct {
		name  string
		p, t  int
		isPT  bool
		label string
	}{
		{"+1/+1", 1, 1, true, "the common grow"},
		{"-1/-1", -1, -1, true, "the common shrink"},
		{"-2/-1", -2, -1, true, "Contagion"},
		{"+1/+0", 1, 0, true, "Dwarven Armorer's power half"},
		{"+0/+1", 0, 1, true, "Dwarven Armorer's toughness half"},
		{"-0/-1", 0, -1, true, "Wall of Roots"},
		{"-1/-0", -1, 0, true, "Jabari's Influence"},
		{"+2/+2", 2, 2, true, "Baron Sengir"},
		{"+1/+2", 1, 2, true, "Armor Thrull"},
		{"-0/-2", 0, -2, true, "Greater Werewolf"},
		{"+10/+10", 10, 10, true, "two digits"},
		{"loyalty", 0, 0, false, "not a P/T counter"},
		{"charge", 0, 0, false, "not a P/T counter"},
		{"", 0, 0, false, "empty"},
		{"1/1", 0, 0, false, "no signs"},
		{"+1/1", 0, 0, false, "toughness sign missing"},
		{"1/+1", 0, 0, false, "power sign missing"},
		{"+X/+X", 0, 0, false, "not digits"},
		{"+1/+1 ", 0, 0, false, "trailing space"},
		{"+1/+1/+1", 0, 0, false, "two slashes"},
		{"+/+1", 0, 0, false, "no digits"},
		{"++1/+1", 0, 0, false, "doubled sign"},
		{"+1000/+1", 0, 0, false, "wider than three digits"},
	}
	for _, c := range cases {
		p, tt, ok := ParsePTCounter(c.name)
		if ok != c.isPT || p != c.p || tt != c.t {
			t.Errorf("%s (%q) = (%d, %d, %v), want (%d, %d, %v)", c.label, c.name, p, tt, ok, c.p, c.t, c.isPT)
		}
	}
}

// ptCreature puts a creature with the given counters on the
// battlefield and returns its ID.
func ptCreature(t *testing.T, g *Game, power, toughness int, counters map[string]int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	owner := g.Seats[0]
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: "Test Creature", TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: owner.ID, Controller: owner.ID,
		Counters: counters,
	})
	return id
}

func ptCard(t *testing.T, g *Game, id uuid.UUID) *Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

func TestEveryPTCounterKindChangesPowerAndToughness(t *testing.T) {
	cases := []struct {
		label    string
		counters map[string]int
		wantP    int
		wantT    int
	}{
		{"-2/-1 on a 3/3 is a 1/2", map[string]int{"-2/-1": 1}, 1, 2},
		{"+1/+0 adds power only", map[string]int{"+1/+0": 1}, 4, 3},
		{"+0/+1 adds toughness only", map[string]int{"+0/+1": 1}, 3, 4},
		{"-0/-1 takes toughness only", map[string]int{"-0/-1": 2}, 3, 1},
		{"mixed kinds sum", map[string]int{"+1/+1": 2, "+1/+0": 1, "-0/-1": 1, "-2/-1": 1, "charge": 5}, 4, 3},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			g := newActiveGame(t)
			id := ptCreature(t, g, 3, 3, c.counters)
			card := ptCard(t, g, id)
			if p := card.PowerForComparison(); p != c.wantP {
				t.Errorf("power %d, want %d", p, c.wantP)
			}
			if tt := card.CurrentToughness(); tt != c.wantT {
				t.Errorf("toughness %d, want %d", tt, c.wantT)
			}
		})
	}
}

// A -0/-1 counter that takes the last point of toughness kills the
// creature (CR 704.5f) — the counter is not decoration.
func TestPTCounterToZeroToughnessKills(t *testing.T) {
	g := newActiveGame(t)
	id := ptCreature(t, g, 0, 1, map[string]int{"-0/-1": 1})
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if ptCard(t, g, id) != nil {
		t.Fatal("a 0/1 with a -0/-1 counter survived the toughness check")
	}
}

// CR 704.5q cancels +1/+1 against -1/-1 and nothing else: a +1/+0 and
// a -1/-0 both stay, and so does a +1/+1 next to a -2/-1.
func TestOnlyPlusOneAndMinusOneAnnihilate(t *testing.T) {
	cases := []struct {
		label    string
		counters map[string]int
		want     map[string]int
	}{
		{"+1/+1 and -1/-1 cancel", map[string]int{"+1/+1": 1, "-1/-1": 1}, nil},
		{"+1/+0 and -1/-0 stay", map[string]int{"+1/+0": 1, "-1/-0": 1}, map[string]int{"+1/+0": 1, "-1/-0": 1}},
		{"+1/+1 and -2/-1 stay", map[string]int{"+1/+1": 1, "-2/-1": 1}, map[string]int{"+1/+1": 1, "-2/-1": 1}},
		{"+0/+1 and -0/-1 stay", map[string]int{"+0/+1": 1, "-0/-1": 1}, map[string]int{"+0/+1": 1, "-0/-1": 1}},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			g := newActiveGame(t)
			id := ptCreature(t, g, 4, 4, c.counters)
			g.WithWriteLock(func() { g.runStateChecksLocked() })
			card := ptCard(t, g, id)
			if card == nil {
				t.Fatal("the creature died")
			}
			if len(card.Counters) != len(c.want) {
				t.Fatalf("counters %v, want %v", card.Counters, c.want)
			}
			for k, v := range c.want {
				if card.Counters[k] != v {
					t.Errorf("%s: %d, want %d (all: %v)", k, card.Counters[k], v, card.Counters)
				}
			}
		})
	}
}

// Undo (Clone) and the persisted snapshot both carry an odd counter
// kind, and the restored creature's P/T reads it.
func TestPTCounterSurvivesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	id := ptCreature(t, g, 3, 3, map[string]int{"-2/-1": 1, "+1/+0": 2})

	clone := g.Clone()
	if c := ptCard(t, clone, id); c == nil || c.PowerForComparison() != 3 || c.CurrentToughness() != 2 {
		t.Fatalf("clone: %+v", c)
	}

	restored, err := throughJSON(t, g.CaptureSnapshot()).Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	c := ptCard(t, restored, id)
	if c == nil {
		t.Fatal("the creature is missing after restore")
	}
	if c.Counters["-2/-1"] != 1 || c.Counters["+1/+0"] != 2 {
		t.Fatalf("restored counters %v", c.Counters)
	}
	if p, tt := c.PowerForComparison(), c.CurrentToughness(); p != 3 || tt != 2 {
		t.Errorf("restored P/T %d/%d, want 3/2", p, tt)
	}
}
