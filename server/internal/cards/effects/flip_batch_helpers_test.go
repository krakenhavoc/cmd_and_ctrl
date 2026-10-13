package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flip_batch_helpers_test.go — what the sixteen follow-on flip cards'
// tests (#3216, ADR 0151) share. The rows and flipCardNow are in
// flip_cards_test.go; each card has its own _test.go file.

// flipMain is a catalog game in the active seat's main phase.
func flipMain(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// unsick clears summoning sickness so a {T} ability can be activated.
func unsick(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

// flipEntered puts a flip card on the battlefield down the real import
// road, ready to use its {T} abilities.
func flipEntered(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := importToBattlefield(t, g, row, p)
	unsick(g, id)
	return id
}

// graveCard puts a card into a player's graveyard.
func graveCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// flipState reads a permanent by instance ID with its layers current.
func flipState(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c := findBattlefieldCardForTest(g, id); c != nil {
			out, found = *c, true
		}
	})
	if !found {
		t.Fatalf("permanent %s is not on the battlefield", id)
	}
	return out
}

// wantFlipped asserts that id is the same permanent, flipped, showing
// the named half and keeping its top half's colours and mana value (CR
// 710.1c).
func wantFlipped(t *testing.T, g *game.Game, id uuid.UUID, name string, colors []string, mv int) game.Card {
	t.Helper()
	c := flipState(t, g, id)
	if !c.Flipped || c.ActiveFace != 1 || c.Name != name {
		t.Fatalf("flipped %v face %d name %q; want flipped as %q", c.Flipped, c.ActiveFace, c.Name, name)
	}
	if !slices.Equal(c.Effective().Colors, colors) || c.ManaValue() != mv {
		t.Errorf("%s: colours %v mana value %d, want %v and %d (CR 710.1c)", name, c.Effective().Colors, c.ManaValue(), colors, mv)
	}
	return c
}

// wantUnflipped asserts that id is still its top half.
func wantUnflipped(t *testing.T, g *game.Game, id uuid.UUID, name string) {
	t.Helper()
	if c := flipState(t, g, id); c.Flipped || c.ActiveFace != 0 || c.Name != name {
		t.Fatalf("flipped %v face %d name %q; want %q unflipped", c.Flipped, c.ActiveFace, c.Name, name)
	}
}

// kiCount is the ki counters on a permanent.
func kiCount(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return flipState(t, g, id).Counters["ki"]
}

// castFlipSpell casts a placeholder spell of the given type line from
// the active seat's hand, as the ki triggers watch for.
func castFlipSpell(t *testing.T, g *game.Game, typeLine string) {
	t.Helper()
	castCatalogSpell(t, g, "Test Spell", typeLine, "", nil)
}

// pendingConfirm reports whether a seat has a yes/no prompt open.
func pendingConfirm(g *game.Game, p *game.Player) bool { return latestConfirmFor(g, p.ID) != nil }

// kiCycleCheck is the whole of the ki flip cards' shared top half
// (Budoka Pupil, Callow Jushi, Cunning Bandit, Faithful Squire, Hired
// Muscle) down the real import road: a counter per Spirit or Arcane
// spell you cast and only those, "you may" on the counter, no flip
// below two counters, a "you may" flip at any player's end step with
// two, and the flip keeping the counters.
func kiCycleCheck(t *testing.T, row func() cards.Card, oracleID, topName, flippedName string, colors []string, mv int) {
	t.Helper()

	t.Run("counters", func(t *testing.T) {
		g, me, _ := flipMain(t)
		id := flipEntered(t, g, row(), me)

		castFlipSpell(t, g, "Instant — Arcane")
		answerLatestTriggerPrompt(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
		if n := kiCount(t, g, id); n != 1 {
			t.Fatalf("after an Arcane spell: %d ki counters, want 1", n)
		}

		castFlipSpell(t, g, "Sorcery")
		passPriorityAroundTable(t, g)
		if n := kiCount(t, g, id); n != 1 || len(g.PendingChoices) != 0 {
			t.Fatalf("a plain sorcery: %d ki counters (pending %d), want 1 and no prompt", n, len(g.PendingChoices))
		}

		castFlipSpell(t, g, "Creature — Spirit")
		answerLatestTriggerPrompt(t, g, me.ID, false)
		passPriorityAroundTable(t, g)
		if n := kiCount(t, g, id); n != 1 {
			t.Fatalf("declining the counter: %d ki counters, want 1", n)
		}

		castFlipSpell(t, g, "Creature — Spirit")
		answerLatestTriggerPrompt(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
		if n := kiCount(t, g, id); n != 2 {
			t.Fatalf("a Spirit spell: %d ki counters, want 2", n)
		}
	})

	t.Run("one counter does not flip", func(t *testing.T) {
		g, me, _ := flipMain(t)
		id := flipEntered(t, g, row(), me)
		pfAddCounters(g, id, "ki", 1)
		advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
		passPriorityAroundTable(t, g)
		if pendingConfirm(g, me) {
			t.Fatal("asked to flip with one ki counter")
		}
		wantUnflipped(t, g, id, topName)
	})

	t.Run("declined", func(t *testing.T) {
		g, me, _ := flipMain(t)
		id := flipEntered(t, g, row(), me)
		pfAddCounters(g, id, "ki", 2)
		advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
		passPriorityAroundTable(t, g)
		answerMayChoice(t, g, me.ID, false)
		passPriorityAroundTable(t, g)
		wantUnflipped(t, g, id, topName)
		if n := kiCount(t, g, id); n != 2 {
			t.Errorf("%d ki counters after declining, want 2", n)
		}
	})

	t.Run("flips at an opponent's end step", func(t *testing.T) {
		g, me, _ := flipMain(t)
		id := flipEntered(t, g, row(), me)
		pfAddCounters(g, id, "ki", 2)
		// "The end step" is every turn's: decline at your own, flip at
		// the next player's.
		advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
		passPriorityAroundTable(t, g)
		answerMayChoice(t, g, me.ID, false)
		passPriorityAroundTable(t, g)
		wantUnflipped(t, g, id, topName)
		advanceToEndStepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
		passPriorityAroundTable(t, g)
		answerMayChoice(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
		c := wantFlipped(t, g, id, flippedName, colors, mv)
		if c.Counters["ki"] != 2 {
			t.Errorf("%d ki counters after the flip, want the same 2", c.Counters["ki"])
		}
		if c.InstanceID != id {
			t.Error("the flip made a new object")
		}
	})
}
