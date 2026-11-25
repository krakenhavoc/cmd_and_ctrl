package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_enter_as_copies_test.go — #6160: ModNextEntersAsCopies, Mystic
// Reflection's "the next time one or more creatures or planeswalkers enter
// this turn, they enter as copies of the chosen creature" (CR 614.1c,
// CR 707.2). The catalog half is in cards/effects.

func pushModelForTest(g *Game, controller uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: "Model Hound", TypeLine: "Creature — Dog", Power: 3, Toughness: 3,
		PrintedPTKnown: true, Owner: controller, Controller: controller,
	})
	return id
}

func registerNextEntersAsCopies(t *testing.T, g *Game, controller, chosen uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.NextCreaturesOrPlaneswalkersEnterAsCopiesThisTurnForEffect(uuid.Nil, controller, chosen, "test: enter as copies") {
			t.Fatal("the replacement was not registered")
		}
	})
}

func battlefieldNameForTest(t *testing.T, g *Game, id uuid.UUID) string {
	t.Helper()
	c, ok := g.battlefieldCardLocked(id)
	if !ok {
		t.Fatalf("%v is not on the battlefield", id)
	}
	return c.Name
}

func moveFromHandForTest(t *testing.T, g *Game, owner, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(id, HandEntryOptions{}); err != nil {
			t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
		}
	})
}

func countNextEntersAsCopies(g *Game) int {
	n := 0
	for i := range g.ScopedEffects {
		if scopedEffectHasMod(&g.ScopedEffects[i], ModNextEntersAsCopies) {
			n++
		}
	}
	return n
}

// The acceptance case: two tokens entering together both become copies of
// the chosen creature, and a creature entering after them does not. A land
// entering first is not a creature or planeswalker and leaves the effect
// waiting.
func TestNextEntersAsCopiesCopiesOneSimultaneousEntry(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	model := pushModelForTest(g, me.ID)
	registerNextEntersAsCopies(t, g, me.ID, model)

	land := uuid.New()
	g.Seats[0].Hand.PushTop(Card{InstanceID: land, Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID})
	moveFromHandForTest(t, g, me.ID, land)
	if got := battlefieldNameForTest(t, g, land); got != "Plains" {
		t.Fatalf("a land entered as %q; only creatures and planeswalkers are copied", got)
	}

	var made []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		made, err = g.CreateTokensForEffect(opp.ID, Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, PrintedPTKnown: true}, 2, TokenEntryOptions{})
		if err != nil {
			t.Fatalf("CreateTokensForEffect: %v", err)
		}
	})
	if len(made) != 2 {
		t.Fatalf("made %d tokens, want 2", len(made))
	}
	for _, id := range made {
		c, _ := g.battlefieldCardLocked(id)
		if c.Name != "Model Hound" || c.Power != 3 || !c.IsToken() || c.Controller != opp.ID {
			t.Errorf("token entered as %q %d/%d token=%v controller=%v; want an opponent's token copy of Model Hound",
				c.Name, c.Power, c.Toughness, c.IsToken(), c.Controller)
		}
	}

	// The same resolution goes on to put another creature onto the
	// battlefield: a separate entry, not part of "the next time".
	later := handCreatureFor(g, me.ID, "Later Bear")
	moveFromHandForTest(t, g, me.ID, later)
	if got := battlefieldNameForTest(t, g, later); got != "Later Bear" {
		t.Errorf("a creature entering after the spent entry became %q", got)
	}

	g.BeginEventBatchForTest()
	if n := countNextEntersAsCopies(g); n != 0 {
		t.Errorf("%d spent records outlived their batch", n)
	}
	last := handCreatureFor(g, me.ID, "Last Bear")
	moveFromHandForTest(t, g, me.ID, last)
	if got := battlefieldNameForTest(t, g, last); got != "Last Bear" {
		t.Errorf("a creature in a later batch became %q", got)
	}
}

// A planeswalker entering becomes the creature (CR 707.2): no loyalty
// counters, since the copy has no loyalty. The chosen creature has died in
// between, so its last-known copiable values are used (CR 608.2h).
func TestNextEntersAsCopiesPlaneswalkerCopiesTheDepartedCreature(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	model := pushModelForTest(g, me.ID)
	registerNextEntersAsCopies(t, g, me.ID, model)
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, model); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}

	pw := uuid.New()
	g.Seats[0].Hand.PushTop(Card{
		InstanceID: pw, Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		StartingLoyalty: 4, StartingLoyaltyKnown: true, Owner: me.ID, Controller: me.ID,
	})
	moveFromHandForTest(t, g, me.ID, pw)
	c, _ := g.battlefieldCardLocked(pw)
	if c.Name != "Model Hound" || !c.IsCreature() || c.IsPlaneswalker() {
		t.Fatalf("the planeswalker entered as %q (%s); want a copy of Model Hound", c.Name, c.TypeLine)
	}
	if n := c.Counters["loyalty"]; n != 0 {
		t.Errorf("the copy entered with %d loyalty counters", n)
	}
}

// The record is a restore point: a snapshot taken while it waits keeps it,
// and the restored game still copies.
func TestNextEntersAsCopiesSurvivesASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	model := pushModelForTest(g, me.ID)
	registerNextEntersAsCopies(t, g, me.ID, model)

	_, restored := roundTrip(t, g)
	if n := countNextEntersAsCopies(restored); n != 1 {
		t.Fatalf("the restored game holds %d records, want 1", n)
	}
	bear := handCreatureFor(restored, restored.Seats[0].ID, "Bear")
	moveFromHandForTest(t, restored, restored.Seats[0].ID, bear)
	if got := battlefieldNameForTest(t, restored, bear); got != "Model Hound" {
		t.Errorf("after a restore the creature entered as %q", got)
	}
}

// "This turn": a record nothing spent ends at cleanup.
func TestNextEntersAsCopiesEndsWithTheTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	model := pushModelForTest(g, me.ID)
	registerNextEntersAsCopies(t, g, me.ID, model)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	bear := handCreatureFor(g, me.ID, "Bear")
	moveFromHandForTest(t, g, me.ID, bear)
	if got := battlefieldNameForTest(t, g, bear); got != "Bear" {
		t.Errorf("a creature entering next turn became %q", got)
	}
}

func TestNextEntersAsCopiesModProblem(t *testing.T) {
	if p := nextEntersAsCopiesModProblem(Mod{Kind: ModNextEntersAsCopies}); p == "" {
		t.Error("a record naming no object was accepted")
	}
	ok := Mod{Kind: ModNextEntersAsCopies, Objects: []ObjectRef{{ID: uuid.New()}}}
	if p := nextEntersAsCopiesModProblem(ok); p != "" {
		t.Errorf("an unspent record was refused: %s", p)
	}
	ok.Objects = append(ok.Objects, ObjectRef{ID: uuid.New()})
	if p := nextEntersAsCopiesModProblem(ok); p == "" {
		t.Error("a spending entry with no batch was accepted")
	}
	ok.SpentBatch = 3
	if p := nextEntersAsCopiesModProblem(ok); p != "" {
		t.Errorf("a spent record was refused: %s", p)
	}
	if p := nextFromSourceModProblem(ok); p != "" {
		t.Errorf("the spent batch was refused as a damage-source field: %s", p)
	}
	if p := blockRequirementModProblem(ok); p != "" {
		t.Errorf("the objects were refused: %s", p)
	}
}
