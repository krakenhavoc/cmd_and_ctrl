package effects

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func ltbEventsFor(g *game.Game, ids ...uuid.UUID) map[uuid.UUID]game.Event {
	wanted := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	got := make(map[uuid.UUID]game.Event, len(ids))
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventLTB && wanted[ev.CardID] {
			got[ev.CardID] = ev
		}
	}
	return got
}

func TestNetherTraitorSeesDeathInLaterInstructionOfSameResolution(t *testing.T) {
	const oracle = "nether-traitor-sequential-deaths-test"
	g := newCatalogGame(t)
	me := g.Seats[0]
	traitor := b12Push(g, me.ID, "Nether Traitor", "Creature — Spirit", drcNetherTraitor, 1, 1)
	victim := drcBear(g, me.ID)
	registerForTest(t, Spec{
		OracleID: oracle, Name: "Two sequential deaths", Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := ctx.Game.DestroyPermanentForEffect(traitor); err != nil {
				return err
			}
			return ctx.Game.DestroyPermanentForEffect(victim)
		},
	})

	castCatalogSpell(t, g, "Two sequential deaths", "Sorcery", oracle, nil)
	drcSettle(t, g, me.ID)

	ltb := ltbEventsFor(g, traitor, victim)
	if len(ltb) != 2 || ltb[traitor].Batch == 0 || ltb[traitor].Batch != ltb[victim].Batch {
		t.Fatalf("sequential LTB events = %#v, want both in the same nonzero resolution batch", ltb)
	}
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("Nether Traitor did not trigger for a later death instruction in the same resolution")
	}
}

func TestNetherTraitorDoesNotSeeCreatureDyingInSameWipeEitherOrder(t *testing.T) {
	for _, traitorFirst := range []bool{true, false} {
		name := "victim first"
		if traitorFirst {
			name = "traitor first"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			var traitor, victim uuid.UUID
			if traitorFirst {
				traitor = b12Push(g, me.ID, "Nether Traitor", "Creature — Spirit", drcNetherTraitor, 1, 1)
				victim = drcBear(g, me.ID)
			} else {
				victim = drcBear(g, me.ID)
				traitor = b12Push(g, me.ID, "Nether Traitor", "Creature — Spirit", drcNetherTraitor, 1, 1)
			}
			castCatalogSpell(t, g, "Damnation", "Sorcery", drcDamnationOracle, nil)
			drcSettle(t, g, me.ID)

			ltb := ltbEventsFor(g, traitor, victim)
			if len(ltb) != 2 || ltb[traitor].Batch == 0 || ltb[traitor].Batch != ltb[victim].Batch {
				t.Fatalf("wipe LTB events = %#v, want both cards in the same nonzero exit batch", ltb)
			}
			if hasPayUnlessFor(g, me.ID) {
				t.Fatal("Nether Traitor saw a creature dying in the same wipe")
			}
		})
	}
}

func TestNetherTraitorDoesNotSeeCreatureInSameSimultaneousSacrificeEitherOrder(t *testing.T) {
	for _, traitorFirst := range []bool{true, false} {
		name := "victim first"
		if traitorFirst {
			name = "traitor first"
		}
		t.Run(name, func(t *testing.T) {
			const oracle = "nether-traitor-simultaneous-sacrifice-test"
			g := newCatalogGame(t)
			me := g.Seats[0]
			var traitor, victim uuid.UUID
			if traitorFirst {
				traitor = b12Push(g, me.ID, "Nether Traitor", "Creature — Spirit", drcNetherTraitor, 1, 1)
				victim = drcBear(g, me.ID)
			} else {
				victim = drcBear(g, me.ID)
				traitor = b12Push(g, me.ID, "Nether Traitor", "Creature — Spirit", drcNetherTraitor, 1, 1)
			}
			registerForTest(t, Spec{
				OracleID: oracle, Name: "Sacrifice both creatures", Completeness: CompletenessFull,
				OnResolve: func(item *game.StackItem, ctx *Context) error {
					ids := []uuid.UUID{traitor, victim}
					if !traitorFirst {
						ids = []uuid.UUID{victim, traitor}
					}
					if n := ctx.Game.SacrificeAllForEffect(item.SourceCardID, ids); n != 2 {
						return fmt.Errorf("SacrificeAllForEffect moved %d cards, want 2", n)
					}
					return nil
				},
			})

			castCatalogSpell(t, g, "Sacrifice both creatures", "Sorcery", oracle, nil)
			drcSettle(t, g, me.ID)
			if !me.Graveyard.Contains(traitor) || !me.Graveyard.Contains(victim) {
				t.Fatal("simultaneous sacrifice did not put both creatures in their owner's graveyard")
			}
			ltb := ltbEventsFor(g, traitor, victim)
			if len(ltb) != 2 || ltb[traitor].Batch == 0 || ltb[traitor].Batch != ltb[victim].Batch {
				t.Fatalf("sacrifice LTB events = %#v, want both cards in the same nonzero exit batch", ltb)
			}
			if hasPayUnlessFor(g, me.ID) {
				t.Fatal("Nether Traitor saw a creature dying in the same simultaneous sacrifice")
			}
		})
	}
}

func TestNetherTraitorInGraveyardSeesLaterSimultaneousBatch(t *testing.T) {
	const oracle = "nether-traitor-later-wipe-test"
	g := newCatalogGame(t)
	me := g.Seats[0]
	traitor := drcTraitorInGraveyard(g, me)
	one, two := drcBear(g, me.ID), drcBear(g, me.ID)
	registerForTest(t, Spec{
		OracleID: oracle, Name: "Destroy two creatures", Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ctx.Game.DestroyPermanentsForEffect([]uuid.UUID{one, two})
			return nil
		},
	})

	castCatalogSpell(t, g, "Destroy two creatures", "Sorcery", oracle, nil)
	drcSettle(t, g, me.ID)
	ltb := ltbEventsFor(g, one, two)
	if len(ltb) != 2 || ltb[one].Batch == 0 || ltb[one].Batch != ltb[two].Batch {
		t.Fatalf("mass-destruction LTB events = %#v, want one simultaneous batch", ltb)
	}
	for i := 0; i < 2; i++ {
		if !hasPayUnlessFor(g, me.ID) {
			t.Fatalf("Nether Traitor did not observe simultaneous death %d", i+1)
		}
		answerPayUnless(t, g, me.ID, false)
		drcSettle(t, g, me.ID)
	}
	if !me.Graveyard.Contains(traitor) {
		t.Fatal("declining the two triggers moved Nether Traitor")
	}
}

func TestNetherTraitorSeesLaterExitFromPublishedBatchContinuation(t *testing.T) {
	const oracle = "nether-traitor-nested-exit-test"
	g := newCatalogGame(t)
	me := g.Seats[0]
	traitor := b12Push(g, me.ID, "Nether Traitor", "Creature — Spirit", drcNetherTraitor, 1, 1)
	other := drcBear(g, me.ID)
	later := drcBear(g, me.ID)
	registerForTest(t, Spec{
		OracleID: oracle, Name: "Exit, then destroy", Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{traitor, other}, hookTail(func(g *game.Game, o game.TailOutcome) error {
				return g.DestroyPermanentForEffect(later)
			}))
		},
	})

	castCatalogSpell(t, g, "Exit, then destroy", "Sorcery", oracle, nil)
	drcSettle(t, g, me.ID)
	ltb := ltbEventsFor(g, traitor, other, later)
	if len(ltb) != 3 || ltb[traitor].Batch == 0 || ltb[traitor].Batch != ltb[other].Batch || ltb[other].Batch != ltb[later].Batch {
		t.Fatalf("continuation LTB events = %#v, want all three in the resolution batch", ltb)
	}
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("later independent exit was suppressed by the still-published outer exit")
	}
}
