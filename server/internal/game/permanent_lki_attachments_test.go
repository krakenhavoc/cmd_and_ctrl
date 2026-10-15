package game

import (
	"encoding/json"
	"testing"
)

// #3780: the record of a departed permanent carries what was attached
// to it, as objects. By the time a sacrifice-cost ability resolves, the
// state-based action has unattached the Equipment, so its own link says
// nothing; the host's record is the only place the answer is left.
func TestPermanentLKIRecordsWhatWasAttachedToIt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	host := pushAttachTestCard(g, me.ID, "Host", "Creature — Soldier")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	other := pushAttachTestCard(g, me.ID, "Other Sword", "Artifact — Equipment")
	hostRef := refOf(t, g, host)
	swordRef := refOf(t, g, sword)
	otherRef := refOf(t, g, other)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})

	live, ok := readPermanent(g, hostRef)
	if !ok || len(live.Attachments) != 1 || live.Attachments[0] != swordRef {
		t.Fatalf("live attachments = %+v, %v; want the one Equipment", live.Attachments, ok)
	}

	destroy(t, g, host)
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if got, _ := battlefieldCardByID(g, sword); got.IsAttached() {
		t.Fatal("setup: the state-based action should have unattached the Equipment")
	}

	gone, ok := readPermanent(g, hostRef)
	if !ok || !gone.Left || len(gone.Attachments) != 1 || gone.Attachments[0] != swordRef {
		t.Fatalf("last-known attachments = %+v, %v; want the Equipment it had", gone.Attachments, ok)
	}
	g.WithWriteLock(func() {
		if live := g.LiveAttachmentsForEffect(gone.Attachments); len(live) != 1 || live[0] != sword {
			t.Errorf("LiveAttachmentsForEffect = %v, want just the Equipment on the battlefield", live)
		}
		stale := []ObjectRef{{ID: sword, Epoch: swordRef.Epoch + 1}, otherRef}
		if live := g.LiveAttachmentsForEffect(stale); len(live) != 1 || live[0] != other {
			t.Errorf("a stale epoch must not count as the same object; got %v", live)
		}
	})
}

// A pinned link that names an EARLIER object of the same card ID is not
// an attachment of this one (#3159, CR 400.7), and is not recorded.
func TestPermanentLKIIgnoresALinkToAnEarlierObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	host := pushAttachTestCard(g, me.ID, "Host", "Creature — Soldier")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
		findBattlefieldCard(g, host).ObjectEpoch++
	})
	hostRef := refOf(t, g, host)
	if info, ok := readPermanent(g, hostRef); !ok || len(info.Attachments) != 0 {
		t.Errorf("attachments = %+v, %v; want none for the new object", info.Attachments, ok)
	}
}

// The attachments ride Clone (undo) and the persisted snapshot with the
// rest of the record, and a clone owns its own backing array.
func TestPermanentLKIAttachmentsRideCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	host := pushAttachTestCard(g, me.ID, "Host", "Creature — Soldier")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	hostRef := refOf(t, g, host)
	swordRef := refOf(t, g, sword)
	g.WithWriteLock(func() { _ = g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: host}) })
	destroy(t, g, host)

	clone := g.Clone()
	if info, ok := readPermanent(clone, hostRef); !ok || len(info.Attachments) != 1 || info.Attachments[0] != swordRef {
		t.Errorf("the clone reads %+v, %v; want the attachment", info.Attachments, ok)
	}
	g.WithWriteLock(func() { g.lastKnownPermanents[host][0].Attachments[0].Epoch = 99 })
	if info, _ := readPermanent(clone, hostRef); len(info.Attachments) != 1 || info.Attachments[0] != swordRef {
		t.Errorf("the clone shares its attachments with the live game: %+v", info.Attachments)
	}

	g.WithWriteLock(func() { g.lastKnownPermanents[host][0].Attachments[0].Epoch = swordRef.Epoch })
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if info, ok := readPermanent(restored, hostRef); !ok || len(info.Attachments) != 1 || info.Attachments[0] != swordRef {
		t.Errorf("the restored game reads %+v, %v; want the attachment", info.Attachments, ok)
	}
}
