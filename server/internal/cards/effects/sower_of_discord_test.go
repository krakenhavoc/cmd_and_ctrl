package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sower_of_discord_test.go — #3389, "as this enters, choose two players".

const sowerOracle = "c6aadaf0-eb47-4228-95e2-48e3ff7b6e2f"

// castSower casts Sower of Discord from the active seat's hand and
// answers both as-enters prompts through the wire resolver.
func castSower(t *testing.T, g *game.Game, first, second uuid.UUID) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := castCatalogSpell(t, g, "Sower of Discord", "Creature — Demon", sowerOracle, nil)
	passPriorityAroundTable(t, g)
	answerChosenPlayer(t, g, active.ID, first)
	answerChosenPlayer(t, g, active.ID, second)
	return id
}

func dealTo(t *testing.T, g *game.Game, source, victim uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, victim, n); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
}

func TestSowerAsksForTwoDifferentPlayers(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	a := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]

	id := castCatalogSpell(t, g, "Sower of Discord", "Creature — Demon", sowerOracle, nil)
	passPriorityAroundTable(t, g)

	c := pendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatalf("no first prompt: %+v", g.PendingChoices)
	}
	if len(c.PickOptions) != len(g.Seats) {
		t.Errorf("every seat is offered first: %d of %d", len(c.PickOptions), len(g.Seats))
	}
	answerChosenPlayer(t, g, active.ID, a.ID)

	c = pendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatalf("no second prompt: %+v", g.PendingChoices)
	}
	if len(c.PickOptions) != len(g.Seats)-1 {
		t.Errorf("the second prompt leaves the first out: %d of %d", len(c.PickOptions), len(g.Seats)-1)
	}
	for _, o := range c.PickOptions {
		if o.Player == a.ID {
			t.Errorf("the first player is offered again")
		}
	}
	answerChosenPlayer(t, g, active.ID, b.ID)

	first, second := g.ChosenPlayersOf(id)
	if first != a.ID || second != b.ID {
		t.Errorf("pair = %v, %v; want %v, %v", first, second, a.ID, b.ID)
	}
}

func TestSowerDrainsTheOtherChosenPlayer(t *testing.T) {
	g := newCatalogGame(t)
	a := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
	castSower(t, g, a.ID, b.ID)
	src := b31Push(g, third.ID, "Some Pinger", "Artifact", "", "{2}", 0, 0)

	la, lb, lt := a.Life, b.Life, third.Life
	dealTo(t, g, src, a.ID, 3)
	if got := la - a.Life; got != 3 {
		t.Errorf("a took %d damage, want 3", got)
	}
	if got := lb - b.Life; got != 3 {
		t.Errorf("b lost %d, want 3", got)
	}

	la, lb = a.Life, b.Life
	dealTo(t, g, src, b.ID, 2)
	if got := la - a.Life; got != 2 {
		t.Errorf("a lost %d, want 2", got)
	}
	if got := lb - b.Life; got != 2 {
		t.Errorf("b took %d, want 2", got)
	}

	la, lb = a.Life, b.Life
	dealTo(t, g, src, third.ID, 4)
	if a.Life != la || b.Life != lb {
		t.Errorf("damage to a third player drained a or b")
	}
	if got := lt - third.Life; got != 4 {
		t.Errorf("third took %d, want 4", got)
	}
}

func TestSowerDrainSurvivesTheSowerLeaving(t *testing.T) {
	g := newCatalogGame(t)
	a := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	id := castSower(t, g, a.ID, b.ID)
	src := b31Push(g, a.ID, "Some Pinger", "Artifact", "", "{2}", 0, 0)

	lb := b.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, a.ID, 5); err != nil {
			t.Fatalf("damage: %v", err)
		}
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := lb - b.Life; got != 5 {
		t.Errorf("b lost %d, want 5 (the trigger was already on the stack)", got)
	}
	if f, s := g.ChosenPlayersOf(id); f != uuid.Nil || s != uuid.Nil {
		t.Errorf("the pair outlived the Sower: %v %v", f, s)
	}
}

func TestSowerPairSurvivesSnapshotAndClone(t *testing.T) {
	g := newCatalogGame(t)
	a := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	id := castSower(t, g, a.ID, b.ID)

	restored := restoreThroughJSON(t, g)
	if f, s := restored.ChosenPlayersOf(id); f != a.ID || s != b.ID {
		t.Errorf("restored pair = %v, %v", f, s)
	}
	cloned := g.Clone()
	if f, s := cloned.ChosenPlayersOf(id); f != a.ID || s != b.ID {
		t.Errorf("cloned pair = %v, %v", f, s)
	}
}

func TestSowerRecastChoosesAgain(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	a := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	id := castSower(t, g, a.ID, b.ID)

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	id2 := castCatalogSpell(t, g, "Sower of Discord", "Creature — Demon", sowerOracle, nil)
	passPriorityAroundTable(t, g)
	answerChosenPlayer(t, g, active.ID, b.ID)
	answerChosenPlayer(t, g, active.ID, active.ID)
	if f, s := g.ChosenPlayersOf(id2); f != b.ID || s != active.ID {
		t.Errorf("recast pair = %v, %v; want %v, %v", f, s, b.ID, active.ID)
	}
}
