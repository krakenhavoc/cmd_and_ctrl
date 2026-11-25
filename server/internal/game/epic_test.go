package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// epic_test.go — CR 702.50, epic (#6141). An uncatalogued fixture: a
// targeted sorcery with epic whose resolution records what it did. The
// catalog card's test is Enduring Ideal's.

const epicFixtureOracle = "epic-fixture-oracle"

// epicRun is one resolution of the fixture: its target, and whether the
// object resolving was a copy.
type epicRun struct {
	target uuid.UUID
	copy   bool
}

// withEpicFixture stubs the catalog with the fixture: "Target player
// loses 1 life. Epic." Every resolution is appended to *runs.
func withEpicFixture(t *testing.T, runs *[]epicRun) {
	t.Helper()
	def := &CardDef{
		Targets: &TargetSpec{Mode: "player", Label: "target player", Players: true, Min: 1, Max: 1},
		Resolve: func(g *Game, item *StackItem) error {
			g.EpicForEffect(item.SourceCardID, "")
			r := epicRun{copy: item.IsCopy}
			if len(item.Targets) == 1 {
				r.target = item.Targets[0].ID
			}
			*runs = append(*runs, r)
			return nil
		},
	}
	prev := CatalogLookup
	t.Cleanup(func() { CatalogLookup = prev })
	CatalogLookup = func(k string) *CardDef {
		if k == epicFixtureOracle {
			return def
		}
		if prev != nil {
			return prev(k)
		}
		return nil
	}
}

func epicFixtureCard(owner uuid.UUID) Card {
	c := NewCard("Epic Fixture", owner)
	c.OracleID = epicFixtureOracle
	c.TypeLine = "Sorcery"
	c.Keywords = []string{"Epic"}
	c.Controller = owner
	return c
}

// epicCastGate asks CastGateLocked whether p may cast a fresh sorcery from
// their hand now.
func epicCastGate(g *Game, p *Player) error {
	c := NewCard("Probe", p.ID)
	c.TypeLine = "Sorcery"
	var err error
	g.WithWriteLock(func() { err = g.CastGateLocked(p.ID, c, ZoneHand, CastSpellParams{}) })
	return err
}

// epicTriggers counts the queued epic triggers.
func epicTriggers(g *Game) int {
	n := 0
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Body == epicCopyBody {
			n++
		}
	}
	return n
}

// The headline (CR 702.50a, 702.50b): once the epic spell resolves its
// controller can't cast spells and nobody else is stopped; the latch and
// the trigger survive a restore; at each of the controller's next two
// upkeeps a copy resolves, offering a new target (CR 707.10c); the copy
// has no epic, so no second trigger is made; and an opponent's upkeep
// copies nothing.
func TestEpicLatchesTheCastBanAndCopiesTheSpellEachUpkeep(t *testing.T) {
	var runs []epicRun
	withEpicFixture(t, &runs)
	g := newActiveGame(t)
	toMainPhase(t, g)
	me, them := g.Seats[g.Turn.ActiveSeat], g.Seats[1-g.Turn.ActiveSeat]
	mySeat := g.Turn.ActiveSeat

	c := epicFixtureCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("cast the epic spell: %v", err)
	}
	if err := epicCastGate(g, me); err != nil {
		t.Fatalf("the ban applies before the spell resolved: %v", err)
	}
	settleStack(t, g)
	if len(runs) != 1 || runs[0].copy || runs[0].target != them.ID {
		t.Fatalf("runs = %+v, want the original once, at them", runs)
	}

	// CR 702.50b: the controller can't cast spells; the opponent can.
	if err := epicCastGate(g, me); !errors.Is(err, ErrCantCast) {
		t.Fatalf("my cast after epic: %v, want ErrCantCast", err)
	}
	if err := epicCastGate(g, them); err != nil {
		t.Fatalf("the opponent's cast is refused: %v", err)
	}
	if n := epicTriggers(g); n != 1 {
		t.Fatalf("%d epic triggers queued, want 1", n)
	}

	// A restore keeps both halves.
	_, g = roundTrip(t, g)
	me, them = g.Seats[mySeat], g.Seats[1-mySeat]
	if err := epicCastGate(g, me); !errors.Is(err, ErrCantCast) {
		t.Fatalf("my cast after the restore: %v, want ErrCantCast", err)
	}
	if n := epicTriggers(g); n != 1 {
		t.Fatalf("%d epic triggers after the restore, want 1", n)
	}

	// The opponent's upkeep is not mine: nothing fires there.
	advanceToSeatStep(t, g, 1-mySeat, StepUpkeep, true)
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Fatalf("the epic trigger fired in the opponent's upkeep")
	}

	for upkeep := 1; upkeep <= 2; upkeep++ {
		advanceToSeatStep(t, g, mySeat, StepUpkeep, true)
		if len(g.PendingTriggers)+len(g.StackMeta) == 0 {
			t.Fatalf("upkeep %d: the epic trigger did not fire", upkeep)
		}
		// The trigger resolves and offers the copy a new target.
		var prompt *PendingChoice
		for i := 0; i < 8 && prompt == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("upkeep %d: pass: %v", upkeep, err)
			}
			prompt = pendingChoiceOfKind(g, PendingChoicePickTarget)
		}
		if prompt == nil {
			t.Fatalf("upkeep %d: no new-target prompt for the copy", upkeep)
		}
		if prompt.Chooser != me.ID {
			t.Fatalf("upkeep %d: the copy's targets are chosen by %s, want me", upkeep, prompt.Chooser)
		}
		want := them.ID
		if upkeep == 2 {
			want = me.ID // choose the other player this time
		}
		if err := g.ResolvePickTargets(prompt.ID, me.ID, []TargetRef{{Kind: TargetPlayer, ID: want}}); err != nil {
			t.Fatalf("upkeep %d: pick the copy's target: %v", upkeep, err)
		}
		settleStack(t, g)
		if len(runs) != upkeep+1 {
			t.Fatalf("upkeep %d: %d runs, want %d", upkeep, len(runs), upkeep+1)
		}
		last := runs[len(runs)-1]
		if !last.copy || last.target != want {
			t.Fatalf("upkeep %d: the run was %+v, want a copy at %s", upkeep, last, want)
		}
		// CR 707.9b: the copy has no epic, so it made no second trigger.
		if n := epicTriggers(g); n != 1 {
			t.Fatalf("upkeep %d: %d epic triggers, want 1", upkeep, n)
		}
		if err := epicCastGate(g, me); !errors.Is(err, ErrCantCast) {
			t.Fatalf("upkeep %d: my cast: %v, want ErrCantCast", upkeep, err)
		}
	}
}

// A plain copy of an epic spell (a Twincast of it) has epic: its own
// resolution sets up a second trigger, so its controller gets two copies
// each upkeep (CR 702.50a, 707.10).
func TestAPlainCopyOfAnEpicSpellHasEpic(t *testing.T) {
	var runs []epicRun
	withEpicFixture(t, &runs)
	g := newActiveGame(t)
	toMainPhase(t, g)
	me, them := g.Seats[g.Turn.ActiveSeat], g.Seats[1-g.Turn.ActiveSeat]

	c := epicFixtureCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(c.InstanceID, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	settleStack(t, g)
	if len(runs) != 2 {
		t.Fatalf("runs = %+v, want the copy and the original", runs)
	}
	if n := epicTriggers(g); n != 2 {
		t.Fatalf("%d epic triggers, want 2: the plain copy has epic", n)
	}
}
