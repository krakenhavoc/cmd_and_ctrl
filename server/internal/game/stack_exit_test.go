package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// stack_exit_test.go — #1318: every exit from the stack retires the
// spell's stack record, whichever verb asked for it, and "exile target
// spell" has its own primitive.
//
// Before #1318 the record was retired only by routes that set
// DropStackMeta — the counter / return-to-hand body and the sandbox
// move. Aang, Swift Savior's airbend of a spell went through the plain
// exile route, moved the card to exile and left a spell record with no
// card behind it. Nothing could resolve past that, and the table
// wedged.

func stackSpellFor(t *testing.T, g *Game, owner *Player, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	pushStackSpell(t, g, Card{
		InstanceID: id, Name: name, TypeLine: "Sorcery",
		Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

// TestEveryStackExitRetiresTheStackRecord runs each effect exit the
// engine exports over a spell on the stack. Each must leave no card on
// the stack AND no StackMeta entry for it — the pair the resolver reads
// together.
func TestEveryStackExitRetiresTheStackRecord(t *testing.T) {
	exits := []struct {
		name string
		run  func(g *Game, id uuid.UUID) error
	}{
		{"ExileCardForEffect (airbend's route)", func(g *Game, id uuid.UUID) error { return g.ExileCardForEffect(id) }},
		{"ExileCardWithPermissionForEffect (airbend)", func(g *Game, id uuid.UUID) error {
			return g.ExileCardWithPermissionForEffect(id, CastPermission{Cost: "{2}", Duration: WhileInZoneDuration(), CastOnly: true})
		}},
		{"ExileCardsForEffect", func(g *Game, id uuid.UUID) error { g.ExileCardsForEffect([]uuid.UUID{id}); return nil }},
		{"ExileSpellForEffect", func(g *Game, id uuid.UUID) error { return g.ExileSpellForEffect(id) }},
		{"BounceToHandForEffect", func(g *Game, id uuid.UUID) error { return g.BounceToHandForEffect(id) }},
		{"TuckToLibraryForEffect", func(g *Game, id uuid.UUID) error { return g.TuckToLibraryForEffect(id, true) }},
		{"PutIntoGraveyardForEffect", func(g *Game, id uuid.UUID) error { return g.PutIntoGraveyardForEffect(id) }},
		{"CounterTargetForEffect", func(g *Game, id uuid.UUID) error { return g.CounterTargetForEffect(id) }},
		{"ReturnSpellToHandForEffect", func(g *Game, id uuid.UUID) error { return g.ReturnSpellToHandForEffect(id) }},
	}
	for _, x := range exits {
		t.Run(x.name, func(t *testing.T) {
			g := newActiveGame(t)
			id := stackSpellFor(t, g, g.Seats[1], "Their Spell")
			g.mu.Lock()
			err := x.run(g, id)
			g.mu.Unlock()
			if err != nil {
				t.Fatalf("exit: %v", err)
			}
			if g.Stack.Contains(id) {
				t.Fatal("the spell is still on the stack")
			}
			if _, ok := g.StackMeta[id]; ok {
				t.Error("the spell left the stack but its StackMeta record stayed behind — the #1318 wedge")
			}
		})
	}
}

// TestAirbentSpellDoesNotWedgeTheTable is the live bug on the deck's
// commander, end to end at the engine: a spell is airbent off the
// stack, and the table then passes priority and moves on to the next
// step. Before #1318 the stale record meant 27 passes never left the
// step.
func TestAirbentSpellDoesNotWedgeTheTable(t *testing.T) {
	g := newActiveGame(t)
	id := stackSpellFor(t, g, g.Seats[1], "Their Spell")
	g.mu.Lock()
	err := g.ExileCardWithPermissionForEffect(id, CastPermission{Cost: "{2}", Duration: WhileInZoneDuration(), CastOnly: true})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("airbend: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the airbent spell is not in exile")
	}
	start := g.Turn.Step
	for i := 0; i < 2*len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		if g.Turn.Step != start {
			return
		}
	}
	t.Fatalf("still in %q after %d passes — the table is wedged on a stale stack record (StackMeta has %d entries)",
		start, 2*len(g.Seats), len(g.StackMeta))
}

// TestExileSpellIsNotACounter: "exile target spell" moves the spell
// without countering it, so a spell that can't be countered is exiled
// all the same and no EventCounterSpell fires.
func TestExileSpellIsNotACounter(t *testing.T) {
	g := newActiveGame(t)
	caster := g.Seats[1]
	id := uuid.New()
	pushStackSpell(t, g, Card{
		InstanceID: id, Name: "Uncounterable Spell", OracleID: "test-cant-be-countered-1318",
		Owner: caster.ID, Controller: caster.ID,
	})
	old := CatalogCantBeCountered
	CatalogCantBeCountered = func(oracleID string) bool { return oracleID == "test-cant-be-countered-1318" }
	t.Cleanup(func() { CatalogCantBeCountered = old })

	var heard []bool
	g.mu.Lock()
	err := g.ExileSpellThenForEffect(id, func(_ *Game, exiled bool) error {
		heard = append(heard, exiled)
		return nil
	})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ExileSpellThenForEffect: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Error("a spell that can't be countered is still exiled — CR 701.6 does not apply")
	}
	if hasEventFor(g, EventCounterSpell, id) {
		t.Error("EventCounterSpell must not fire — this is not a counter")
	}
	if len(heard) != 1 || !heard[0] {
		t.Errorf("then heard %v, want [true]", heard)
	}
}

// TestExileSpellRefusesANonSpell: the fire-and-forget form says spell
// and means it; the Then form treats a spell that has already gone as
// "nothing happened" and still answers its caller.
func TestExileSpellRefusesANonSpell(t *testing.T) {
	g := newActiveGame(t)
	bear := pushBear(g, g.Seats[0].ID)
	g.mu.Lock()
	err := g.ExileSpellForEffect(bear)
	var heard []bool
	thenErr := g.ExileSpellThenForEffect(bear, func(_ *Game, exiled bool) error {
		heard = append(heard, exiled)
		return nil
	})
	g.mu.Unlock()
	if !errors.Is(err, ErrCardNotOnStack) {
		t.Errorf("ExileSpellForEffect on a permanent = %v, want ErrCardNotOnStack", err)
	}
	if thenErr != nil {
		t.Errorf("ExileSpellThenForEffect on a permanent = %v, want nil", thenErr)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the permanent moved")
	}
	if len(heard) != 1 || heard[0] {
		t.Errorf("then heard %v, want [false]", heard)
	}
}

// TestExiledCommanderSpellWaitsForTheAnswer: the exile of a commander
// spell pauses on CR 903.9 with the record still in place (the spell
// has not moved), and `then` hears the settled outcome — false when the
// owner takes the command zone, because nothing reached exile to plot.
func TestExiledCommanderSpellWaitsForTheAnswer(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[1]
	cmdID := seatCommander(t, g.Stack, owner)
	top, _ := g.Stack.Top()
	_, _ = g.Stack.Remove(cmdID)
	pushStackSpell(t, g, top)

	var heard []bool
	g.mu.Lock()
	err := g.ExileSpellThenForEffect(cmdID, func(_ *Game, exiled bool) error {
		heard = append(heard, exiled)
		return nil
	})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ExileSpellThenForEffect: %v", err)
	}
	if len(heard) != 0 {
		t.Fatalf("then ran before the CR 903.9 answer: %v", heard)
	}
	if _, ok := g.StackMeta[cmdID]; !ok {
		t.Fatal("the record went before the spell did")
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, g.Exile, g.Stack)
	if _, ok := g.StackMeta[cmdID]; ok {
		t.Error("the record outlived the commander's move to the command zone")
	}
	if len(heard) != 1 || heard[0] {
		t.Errorf("then heard %v, want [false] — the commander went home, not to exile", heard)
	}
}
