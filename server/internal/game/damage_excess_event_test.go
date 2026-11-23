package game

import (
	"testing"

	"github.com/google/uuid"
)

// damage_excess_event_test.go — #5908: the damage event carries how much
// of it was excess (CR 120.10), and the turn's damage record keeps what
// kind of source dealt it (Aegar, the Freezing Flame).

// lastDealDamageEventTo is the newest EventDealDamage to target this turn.
func lastDealDamageEventTo(t *testing.T, g *Game, target uuid.UUID) Event {
	t.Helper()
	var out Event
	found := false
	g.ReadSnapshot(func() {
		for _, ev := range g.EventsThisTurn() {
			if ev.Kind == EventDealDamage && ev.Target == target {
				out, found = ev, true
			}
		}
	})
	if !found {
		t.Fatal("no damage event to the target")
	}
	return out
}

func TestDamageEventCarriesTheExcessAmount(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := excessSourceForTest(g, me.ID)
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)
	dealNoncombatForTest(t, g, src, victim, 5)
	if got := lastDealDamageEventTo(t, g, victim).ExcessDamage; got != 3 {
		t.Errorf("excess on 5 damage to a 2/2 = %d, want 3", got)
	}
	exact := pushScopedTestCreature(g, opp.ID, 3, 3)
	dealNoncombatForTest(t, g, src, exact, 3)
	if got := lastDealDamageEventTo(t, g, exact).ExcessDamage; got != 0 {
		t.Errorf("excess on exactly lethal damage = %d, want 0", got)
	}
	// Damage already marked counts toward lethal (CR 120.6).
	marked := pushScopedTestCreature(g, opp.ID, 3, 3)
	dealNoncombatForTest(t, g, src, marked, 2)
	dealNoncombatForTest(t, g, src, marked, 2)
	if got := lastDealDamageEventTo(t, g, marked).ExcessDamage; got != 1 {
		t.Errorf("excess on the second 2 damage to a 3/3 = %d, want 1", got)
	}
}

func TestDamageRecordKeepsTheSourceKind(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wizard := excessSourceForTest(g, me.ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == wizard {
				g.Battlefield.Cards[i].TypeLine = "Creature — Human Wizard"
			}
		}
	})
	g.layerVersion.Add(1)
	bear := excessSourceForTest(g, me.ID)
	victim := pushScopedTestCreature(g, opp.ID, 9, 9)
	other := pushScopedTestCreature(g, opp.ID, 9, 9)
	dealNoncombatForTest(t, g, wizard, victim, 1)
	dealNoncombatForTest(t, g, bear, other, 1)
	var wizardHit, bearHit, wrongController bool
	g.WithWriteLock(func() {
		wizardHit = g.PermanentDealtDamageThisTurnByKindForEffect(victim, me.ID, true, "Giant", "Wizard")
		bearHit = g.PermanentDealtDamageThisTurnByKindForEffect(other, me.ID, true, "Giant", "Wizard")
		wrongController = g.PermanentDealtDamageThisTurnByKindForEffect(victim, opp.ID, true, "Giant", "Wizard")
	})
	if !wizardHit {
		t.Error("damage from a Wizard you controlled was not found")
	}
	if bearHit {
		t.Error("damage from a creature that is neither a Giant, a Wizard nor a spell was found")
	}
	if wrongController {
		t.Error("the damage was read for a player who did not control the source")
	}
}

// The intervening "if" of Aegar only grows within a turn: it reads nothing
// while it holds, and is refused before any such damage (another item of
// the batch could deal it) or with no triggering event to name the victim.
func TestDamagedByGiantWizardOrSpellConditionReadsNothingWhileItHolds(t *testing.T) {
	f := newCondFixture(t)
	victim := pushScopedTestCreature(f.g, f.opp.ID, 9, 9)
	item := f.item()
	item.Trigger = &TriggerContext{Event: Event{Kind: EventDealDamage, Target: victim}}
	step := cond(FootprintCondDamagedByGiantWizardOrSpell, 0)
	check := func() bool {
		a := newTriggerAccess(item.Controller)
		a.source = item.SourceCardID
		var ok bool
		f.g.WithWriteLock(func() { ok = f.g.addFootprintStepLocked(&a, item, step, nil) })
		return ok
	}
	if check() {
		t.Error("before any damage the condition is false, and the step is refused")
	}
	// A Wizard you control damages it.
	wizard := excessSourceForTest(f.g, f.me.ID)
	f.g.WithWriteLock(func() {
		for i := range f.g.Battlefield.Cards {
			if f.g.Battlefield.Cards[i].InstanceID == wizard {
				f.g.Battlefield.Cards[i].TypeLine = "Creature — Human Wizard"
			}
		}
	})
	f.g.layerVersion.Add(1)
	dealNoncombatForTest(t, f.g, wizard, victim, 1)
	if !check() {
		t.Error("with a Wizard's damage on record the condition holds and the step reads nothing")
	}
	item.Trigger = nil
	if check() {
		t.Error("with no triggering event the step is refused")
	}
}

// "If a +1/+1 counter was put on a permanent under your control this turn"
// only becomes true within a turn: it reads nothing while it holds and is
// refused before any counter is put.
func TestPlusOneCounterConditionReadsNothingWhileItHolds(t *testing.T) {
	f := newCondFixture(t)
	step := cond(FootprintCondPlusOneCounterPutOnYourPermanent, 0)
	if _, ok := f.access(t, step); ok {
		t.Error("before any counter is put the condition is false, and the step is refused")
	}
	mine := pushScopedTestCreature(f.g, f.me.ID, 2, 2)
	theirs := pushScopedTestCreature(f.g, f.opp.ID, 2, 2)
	f.g.WithWriteLock(func() {
		if err := f.g.AddCounterForEffect(theirs, CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	if _, ok := f.access(t, step); ok {
		t.Error("a counter on an opponent's permanent satisfied the condition")
	}
	f.g.WithWriteLock(func() {
		if err := f.g.AddCounterForEffect(mine, CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	a := f.mustAccess(t, step)
	if a.boardReads || a.monarchRead {
		t.Errorf("read the board or the monarch: %+v", a)
	}
}
