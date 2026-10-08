package game

import (
	"testing"

	"github.com/google/uuid"
)

// #2621 / CR 603.2, 603.10: a permanent that has left the battlefield
// does not trigger on an event that happens afterwards. The open
// simultaneous exit is a window for leaves-the-battlefield watchers
// (CR 603.10a), not for an "another creature enters" watcher that is
// already gone.
func TestDepartedWatcherDoesNotTriggerOnALaterEntry(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	installSuppressionCatalog(t)
	watcher := pushBattlefieldForTest(g, me.ID, "Marwyn", "Creature — Elf", suppWatcher)
	elf := suppInHand(me, "Elf", "Creature — Elf", "")

	g.WithWriteLock(func() {
		closeBatch := g.beginSimultaneousExitLocked([]uuid.UUID{watcher})
		defer closeBatch()
		left, err := g.Battlefield.Remove(watcher)
		if err != nil {
			t.Fatalf("remove the watcher: %v", err)
		}
		me.Graveyard.PushTop(left)
		if _, err := g.PutFromHandOntoBattlefieldForEffect(elf, HandEntryOptions{}); err != nil {
			t.Fatalf("put from hand: %v", err)
		}
	})
	if got := suppQueued(g, "watch"); got != 0 {
		t.Fatalf("a watcher that had left triggered on a later entry: %d instances, want 0", got)
	}
}

// The simultaneous case stays: a dies-watcher wiped alongside the
// creature it watches still sees that death (CR 603.10a).
func TestDepartedWatcherStillSeesASimultaneousDeath(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	installSuppressionCatalog(t)
	watcher := pushBattlefieldForTest(g, me.ID, "Watcher", "Creature — Elf", suppDiesAny)
	other := pushBattlefieldForTest(g, me.ID, "Other", "Creature — Elf", "")

	g.WithWriteLock(func() {
		g.DestroyPermanentsForEffect([]uuid.UUID{watcher, other})
	})
	if got := suppQueued(g, "dies watch"); got != 1 {
		t.Fatalf("wiped watcher saw %d deaths of the other creature, want 1", got)
	}
}
