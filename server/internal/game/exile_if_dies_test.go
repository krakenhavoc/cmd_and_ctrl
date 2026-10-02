package game

import (
	"testing"

	"github.com/google/uuid"
)

// exile_if_dies_test.go pins the engine half of ADR 0108 §1 and §2
// (#1886, #1887): the exileIfWouldDie replacement, its pinned and
// scoped forms, the cantBeRegenerated gate, and the per-recipient damage
// continuation. The card behaviour is pinned in cards/effects.

func zoneOf(t *testing.T, g *Game, id uuid.UUID) ZoneKind {
	t.Helper()
	var kind ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

func destroyForTest(t *testing.T, g *Game, id uuid.UUID, opts ...DestroyOptions) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id, opts...); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}

// CR 614.1a, 700.4: "if it would die this turn, exile it instead"
// replaces the move to the graveyard, by any route that is a death —
// destruction, sacrifice, the zero-toughness state-based action.
func TestExileIfItWouldDieReplacesEveryDeath(t *testing.T) {
	for _, tc := range []struct {
		name string
		kill func(t *testing.T, g *Game, id uuid.UUID)
	}{
		{"destroyed", func(t *testing.T, g *Game, id uuid.UUID) { destroyForTest(t, g, id) }},
		{"sacrificed", func(t *testing.T, g *Game, id uuid.UUID) {
			g.WithWriteLock(func() {
				if err := g.SacrificePermanentForEffect(id); err != nil {
					t.Fatalf("sacrifice: %v", err)
				}
			})
		}},
		{"zero toughness", func(t *testing.T, g *Game, id uuid.UUID) {
			g.WithWriteLock(func() {
				g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id),
					[]Mod{ModifyPTMod(-5, -5)}, g.UntilEndOfTurnDuration(), "shrink")
				g.runStateChecksLocked()
			})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			bear := pushScopedTestCreature(g, g.Seats[1].ID, 2, 2)
			var ok bool
			g.WithWriteLock(func() { ok = g.ExileIfItWouldDieThisTurnForEffect(uuid.Nil, bear, g.Seats[0].ID, "Lava Coil") })
			if !ok {
				t.Fatal("no record written")
			}
			tc.kill(t, g, bear)
			if z := zoneOf(t, g, bear); z != ZoneExile {
				t.Errorf("the creature went to %q, want exile", z)
			}
		})
	}
}

// Leaving the battlefield another way is not dying: a bounce is
// untouched.
func TestExileIfItWouldDieLeavesABounceAlone(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[1].ID, 2, 2)
	g.WithWriteLock(func() {
		g.ExileIfItWouldDieThisTurnForEffect(uuid.Nil, bear, uuid.Nil, "Lava Coil")
		g.BounceCardsToHandForEffect([]uuid.UUID{bear})
	})
	if z := zoneOf(t, g, bear); z != ZoneHand {
		t.Errorf("the bounced creature went to %q, want hand", z)
	}
}

// CR 400.7: the record names the object. A creature that left and came
// back is a new object, and dies normally.
func TestExileIfItWouldDieDoesNotFollowAFlicker(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[1].ID, 2, 2)
	g.WithWriteLock(func() {
		g.ExileIfItWouldDieThisTurnForEffect(uuid.Nil, bear, uuid.Nil, "Lava Coil")
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].EnteredBattlefieldAt += 1_000_000
			}
		}
	})
	destroyForTest(t, g, bear)
	if z := zoneOf(t, g, bear); z != ZoneGraveyard {
		t.Errorf("the new object went to %q, want the graveyard", z)
	}
}

// CR 514.2: "this turn" ends at cleanup.
func TestExileIfItWouldDieEndsAtCleanup(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[1].ID, 2, 2)
	g.WithWriteLock(func() {
		g.ExileIfItWouldDieThisTurnForEffect(uuid.Nil, bear, uuid.Nil, "Lava Coil")
		g.sweepTurnEndLocked()
	})
	destroyForTest(t, g, bear)
	if z := zoneOf(t, g, bear); z != ZoneGraveyard {
		t.Errorf("next turn the creature went to %q, want the graveyard", z)
	}
}

// CR 611.2c: "If a creature would die this turn" reads its set live, so
// a creature that entered after the spell resolved is exiled too, and a
// permanent that is not a creature is not.
func TestExileIfCreaturesWouldDieReadsItsSetLive(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	var ok bool
	g.WithWriteLock(func() {
		ok = g.ExileIfCreaturesWouldDieThisTurnForEffect(uuid.Nil, me, ScopeCreatures, "Flaying Tendrils")
	})
	if !ok {
		t.Fatal("no record written")
	}
	mine := pushScopedTestCreature(g, me, 2, 2)
	theirs := pushScopedTestCreature(g, opp, 2, 2)
	rock := uuid.New()
	g.Battlefield.PushTop(Card{InstanceID: rock, Name: "Mind Stone", TypeLine: "Artifact", Owner: opp, Controller: opp})
	for _, id := range []uuid.UUID{mine, theirs, rock} {
		destroyForTest(t, g, id)
	}
	for id, want := range map[uuid.UUID]ZoneKind{mine: ZoneExile, theirs: ZoneExile, rock: ZoneGraveyard} {
		if z := zoneOf(t, g, id); z != want {
			t.Errorf("%s went to %q, want %q", id, z, want)
		}
	}
	g.ReadSnapshot(func() {
		if got := g.ExileIfCreaturesWouldDieThisTurnLabels(); len(got) != 1 || got[0] != "Flaying Tendrils" {
			t.Errorf("banner labels %v, want [Flaying Tendrils]", got)
		}
	})
}

// "If a creature an opponent controls would die this turn" (Malicious
// Eclipse): the controller is read as the creature would die.
func TestExileIfOpponentsCreaturesWouldDie(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	g.WithWriteLock(func() {
		g.ExileIfCreaturesWouldDieThisTurnForEffect(uuid.Nil, me, ScopeOpponentsCreatures, "Malicious Eclipse")
	})
	mine := pushScopedTestCreature(g, me, 2, 2)
	theirs := pushScopedTestCreature(g, opp, 2, 2)
	destroyForTest(t, g, mine)
	destroyForTest(t, g, theirs)
	if z := zoneOf(t, g, mine); z != ZoneGraveyard {
		t.Errorf("my creature went to %q, want the graveyard", z)
	}
	if z := zoneOf(t, g, theirs); z != ZoneExile {
		t.Errorf("the opponent's creature went to %q, want exile", z)
	}
}

// Two records on one dying creature are one modification: no CR 616
// prompt is asked, and it is exiled.
func TestTwoExileIfDiesRecordsAskNothing(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[1].ID, 2, 2)
	g.WithWriteLock(func() {
		g.ExileIfItWouldDieThisTurnForEffect(uuid.Nil, bear, g.Seats[0].ID, "Lava Coil")
		g.ExileIfItWouldDieThisTurnForEffect(uuid.Nil, bear, g.Seats[1].ID, "Magma Spray")
	})
	g.ReadSnapshot(func() {
		if got := g.ExileIfItWouldDieLabels(bear); len(got) != 2 {
			t.Errorf("chip labels %v, want both records", got)
		}
	})
	destroyForTest(t, g, bear)
	if n := len(g.PendingChoices); n != 0 {
		t.Fatalf("%d prompts pending, want none", n)
	}
	if z := zoneOf(t, g, bear); z != ZoneExile {
		t.Errorf("the creature went to %q, want exile", z)
	}
}

// --- ADR 0108 §2: can't be regenerated this turn ---------------------

// CR 701.19c: a creature that can't be regenerated this turn is
// destroyed through its regeneration shield.
func TestCantBeRegeneratedThisTurnIgnoresTheShield(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[1].ID, 2, 2)
	var ok bool
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(bear); err != nil {
			t.Fatalf("regenerate: %v", err)
		}
		ok = g.CantBeRegeneratedThisTurnForEffect(uuid.Nil, bear, "Incinerate")
	})
	if !ok {
		t.Fatal("no record written")
	}
	g.ReadSnapshot(func() {
		if !g.PermanentCantBeRegeneratedForEffect(bear) {
			t.Error("the chip does not say it can't be regenerated")
		}
	})
	destroyForTest(t, g, bear)
	if z := zoneOf(t, g, bear); z != ZoneGraveyard {
		t.Errorf("the creature went to %q, want the graveyard", z)
	}
}

// The control: without the mark, the shield saves it. And the mark ends
// at cleanup, with the turn.
func TestCantBeRegeneratedEndsAtCleanup(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[1].ID, 2, 2)
	g.WithWriteLock(func() {
		g.CantBeRegeneratedThisTurnForEffect(uuid.Nil, bear, "Incinerate")
		g.sweepTurnEndLocked()
		if err := g.RegenerateForEffect(bear); err != nil {
			t.Fatalf("regenerate: %v", err)
		}
	})
	destroyForTest(t, g, bear)
	if z := zoneOf(t, g, bear); z != ZoneBattlefield {
		t.Errorf("next turn the shield did not save it: it went to %q", z)
	}
}

// CR 701.19c: the shield is not applied, so it is not used up either —
// and with "exile it instead" beside it (Disintegrate), the creature is
// exiled and the shield was never an ordering question.
func TestCantBeRegeneratedDoesNotSpendTheShield(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[1].ID, 2, 2)
	destroy := &ReplacementEvent{Kind: RepEventMove, CardID: bear, OldZone: ZoneBattlefield, NewZone: ZoneGraveyard, Destruction: true}
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(bear); err != nil {
			t.Fatalf("regenerate: %v", err)
		}
		g.CantBeRegeneratedThisTurnForEffect(uuid.Nil, bear, "Disintegrate")
		g.ExileIfItWouldDieThisTurnForEffect(uuid.Nil, bear, uuid.Nil, "Disintegrate")
		if g.regenerationShieldAppliesLocked(destroy) {
			t.Error("the shield still applies")
		}
		if n := g.RegenerationShieldsOn(bear); n != 1 {
			t.Errorf("%d shields after the refusal, want 1", n)
		}
	})
	destroyForTest(t, g, bear)
	if n := len(g.PendingChoices); n != 0 {
		t.Fatalf("%d prompts pending, want none", n)
	}
	if z := zoneOf(t, g, bear); z != ZoneExile {
		t.Errorf("the creature went to %q, want exile", z)
	}
}

// CR 701.19b: a static regeneration asks the same gate.
func TestStaticRegenerationAsksTheGate(t *testing.T) {
	g := newActiveGame(t)
	bear := pushScopedTestCreature(g, g.Seats[1].ID, 2, 2)
	destroy := &ReplacementEvent{Kind: RepEventMove, CardID: bear, OldZone: ZoneBattlefield, NewZone: ZoneGraveyard, Destruction: true}
	g.WithWriteLock(func() {
		if !g.RegenerationAllowedForEffect(destroy, bear) {
			t.Error("an unmarked destruction may not be regenerated")
		}
		sac := *destroy
		sac.Destruction = false
		if g.RegenerationAllowedForEffect(&sac, bear) {
			t.Error("a sacrifice may be regenerated")
		}
		rider := *destroy
		rider.CantBeRegenerated = true
		if g.RegenerationAllowedForEffect(&rider, bear) {
			t.Error("Wrath's rider is ignored")
		}
		g.CantBeRegeneratedThisTurnForEffect(uuid.Nil, bear, "Incinerate")
		if g.RegenerationAllowedForEffect(destroy, bear) {
			t.Error("the turn's mark is ignored")
		}
	})
}

// --- ADR 0108 §1 decision 4: damage per recipient --------------------

func TestDealDamageEachEachTellsEachRecipient(t *testing.T) {
	g := newActiveGame(t)
	a := pushScopedTestCreature(g, g.Seats[1].ID, 2, 4)
	b := pushScopedTestCreature(g, g.Seats[1].ID, 2, 4)
	gone := uuid.New()
	victim := g.Seats[0].ID
	got := map[uuid.UUID]int{}
	total := -1
	g.WithWriteLock(func() {
		g.PreventNextDamageThisTurnForEffect(uuid.Nil, b, 2, false, "shield")
		err := g.DealDamageEachEachThenForEffect(uuid.New(), []uuid.UUID{a, gone, b, victim}, 3,
			func(_ *Game, target uuid.UUID, dealt int) error {
				got[target] = dealt
				return nil
			},
			func(_ *Game, n int) error {
				total = n
				return nil
			})
		if err != nil {
			t.Fatalf("deal: %v", err)
		}
	})
	if got[a] != 3 || got[b] != 1 || got[victim] != 3 {
		t.Errorf("per recipient %v, want a=3 b=1 victim=3", got)
	}
	if _, ok := got[gone]; ok {
		t.Error("a recipient that is gone was reported")
	}
	if total != 7 {
		t.Errorf("total %d, want 7", total)
	}
}
