package game

import (
	"testing"

	"github.com/google/uuid"
)

// trigger_only_once_test.go — the engine half of "This ability triggers
// only once" (#5245, TriggeredAbility.OnlyOnce). The card half is
// effects/once_ever_skip_draw_cards_test.go (Acrobatic Cheerleader).

const onlyOnceOracle, onlyOnceKey = "only-once-source", "Only Once — the ability"

// onlyOnceFixture puts a source on the battlefield whose one trigger
// watches EventETB, matches any entry, and triggers only once.
func onlyOnceFixture(t *testing.T) (*Game, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	src := seedTriggerSource(g, g.Seats[0], onlyOnceOracle)
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != onlyOnceOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches:   []EventKind{EventETB},
			Key:       onlyOnceKey,
			OnlyOnce:  true,
			AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, onlyOnceKey)
			},
		}}
	})
	return g, src
}

// onlyOnceFire emits one entry and reports how many instances of the
// ability it queued. The queue is emptied first, so each call counts
// only its own.
func onlyOnceFire(g *Game) int {
	n := 0
	g.WithWriteLock(func() {
		g.PendingTriggers = nil
		g.EmitEvent(Event{Kind: EventETB, CardID: uuid.New()})
		for _, it := range g.PendingTriggers {
			if it.Label == onlyOnceKey {
				n++
			}
		}
		g.PendingTriggers = nil
	})
	return n
}

// The first match triggers and spends the ability; the second, on the
// same turn and on a later one, does not.
func TestTriggersOnlyOnceFiresTheFirstTimeAndNeverAgain(t *testing.T) {
	g, src := onlyOnceFixture(t)
	if got := onlyOnceFire(g); got != 1 {
		t.Fatalf("first match queued %d instances, want 1", got)
	}
	if !g.TriggeredOnlyOnceSpent(src, onlyOnceKey) {
		t.Error("the record is not written after the first trigger")
	}
	if got := onlyOnceFire(g); got != 0 {
		t.Errorf("second match on the same turn queued %d instances, want 0", got)
	}

	seq := g.Turn.Seq
	for i := 0; i < 40 && g.Turn.Seq == seq; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Seq == seq {
		t.Fatal("the turn never advanced")
	}
	if got := onlyOnceFire(g); got != 0 {
		t.Errorf("a match on a later turn queued %d instances, want 0", got)
	}
}

// A zone change makes a new object (CR 400.7), which reads a key nothing
// has written and may trigger again, once.
func TestTriggersOnlyOnceIsPerObject(t *testing.T) {
	g, src := onlyOnceFixture(t)
	if got := onlyOnceFire(g); got != 1 {
		t.Fatalf("first match queued %d instances, want 1", got)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == src {
				// The epoch is what MoveCard bumps on a zone change.
				g.Battlefield.Cards[i].ObjectEpoch++
			}
		}
	})
	if g.TriggeredOnlyOnceSpent(src, onlyOnceKey) {
		t.Error("the new object reads the old object's record")
	}
	if got := onlyOnceFire(g); got != 1 {
		t.Errorf("the new object queued %d instances, want 1", got)
	}
	if got := onlyOnceFire(g); got != 0 {
		t.Errorf("the new object's second match queued %d instances, want 0", got)
	}
}

// A restore keeps the ability spent: the record is in the snapshot.
func TestTriggersOnlyOnceSurvivesARestore(t *testing.T) {
	g, src := onlyOnceFixture(t)
	if got := onlyOnceFire(g); got != 1 {
		t.Fatalf("first match queued %d instances, want 1", got)
	}
	_, restored := roundTrip(t, g)
	if !restored.TriggeredOnlyOnceSpent(src, onlyOnceKey) {
		t.Error("the restored game forgot the record")
	}
	if got := onlyOnceFire(restored); got != 0 {
		t.Errorf("after a restore a match queued %d instances, want 0", got)
	}
}

// "If an effect would cause [the] ability to trigger again, that part of
// the effect won't do anything" (the Acrobatic Cheerleader ruling): a
// doubler adds no second instance. An ordinary trigger beside it, on the
// same event, is still doubled, so the doubler does apply.
func TestTriggersOnlyOnceIsNotDoubled(t *testing.T) {
	g, _ := onlyOnceFixture(t)
	const plainOracle, plainKey, doublerOracle = "only-once-plain", "Plain — the ability", "only-once-doubler"
	seedTriggerSource(g, g.Seats[0], plainOracle)
	onlyOnce := CatalogTriggers
	withCatalogTriggers(t, func(key string) []TriggeredAbility {
		if key != plainOracle {
			return onlyOnce(key)
		}
		return []TriggeredAbility{{
			Watches:   []EventKind{EventETB},
			Key:       plainKey,
			AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, plainKey)
			},
		}}
	})
	pushBattlefieldForTest(g, g.Seats[0].ID, "Doubler", "Artifact", doublerOracle)
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key != doublerOracle {
			return nil
		}
		return []TriggerDoubler{{Label: "Doubler", Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
	})
	counts := map[string]int{}
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: uuid.New()})
		for _, it := range g.PendingTriggers {
			counts[it.Label]++
		}
	})
	if counts[onlyOnceKey] != 1 {
		t.Errorf("under a doubler the only-once ability queued %d instances, want 1", counts[onlyOnceKey])
	}
	if counts[plainKey] != 2 {
		t.Errorf("the ordinary ability queued %d instances, want 2 (the doubler did not apply)", counts[plainKey])
	}
}
