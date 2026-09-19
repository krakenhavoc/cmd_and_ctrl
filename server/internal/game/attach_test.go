package game

import (
	"testing"

	"github.com/google/uuid"
)

// attach_test.go pins the S24 attachment relation (ADR 0036): the
// primitives, the CR 704.5m/n state-based action's two opposite
// branches, the CR 400.7 clear on battlefield exit, the CR 613.7d
// timestamp, and the undo/clone carry that clone.go gets for free
// and therefore has nothing enforcing it.

// pushAttachTestCard seeds a battlefield card and fires the zone-move
// event so the layer listener stamps the CR 613 timestamp — the
// attachment SBA runs after a recompute, so a card the engine has
// never seen enter would read a nil effective cache.
func pushAttachTestCard(g *Game, controller uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Owner:      controller,
		Controller: controller,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  id,
			OldZone: ZoneHand,
			NewZone: ZoneBattlefield,
		})
	})
	return id
}

// battlefieldCardByID returns a copy of the named battlefield card,
// or ok=false when it is no longer there.
func battlefieldCardByID(g *Game, id uuid.UUID) (Card, bool) {
	var out Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c, true
				return
			}
		}
	})
	return out, found
}

func graveyardHas(p *Player, id uuid.UUID) bool {
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestAttachForEffectLinksAndEmits(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")

	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})

	got, ok := battlefieldCardByID(g, sword)
	if !ok {
		t.Fatal("sword left the battlefield")
	}
	if !got.IsAttachedTo(bear) {
		t.Fatalf("AttachedTo = %+v, want card %s", got.AttachedTo, bear)
	}
	if got.AttachedAt == 0 {
		t.Error("AttachedAt not stamped — CR 613.7d timestamp is missing")
	}

	var attached *Event
	g.ReadSnapshot(func() {
		for i := range g.Events {
			if g.Events[i].Kind == EventAttach {
				ev := g.Events[i]
				attached = &ev
			}
		}
	})
	if attached == nil {
		t.Fatal("no EventAttach emitted")
	}
	if attached.CardID != sword || attached.Target != bear {
		t.Errorf("EventAttach{CardID:%s, Target:%s}, want {%s, %s}",
			attached.CardID, attached.Target, sword, bear)
	}

	// Reverse lookup, which is the only direction the client gets.
	var rev []uuid.UUID
	g.ReadSnapshot(func() { rev = g.AttachmentsOf(bear) })
	if len(rev) != 1 || rev[0] != sword {
		t.Errorf("AttachmentsOf(bear) = %v, want [%s]", rev, sword)
	}
}

// CR 701.3b: an attach that cannot happen DOES NOTHING. Three ways it
// cannot — the host is gone, the attachment is gone, the two are the
// same permanent — and all three are a quiet EventAttachSkipped rather
// than an error, because the ability that asked has resolved and an
// error would be reported as a bug (#812). Only a TargetRef that names
// nothing at all is still a caller error.
func TestAttachForEffectSkipsQuietlyWhenItCannotHappen(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")

	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")

	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: sword}); err != nil {
			t.Errorf("self-attach: got %v, want nil", err)
		}
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: uuid.New()}); err != nil {
			t.Errorf("absent host: got %v, want nil", err)
		}
		if err := g.AttachForEffect(sword, TargetRef{}); err != ErrInvalidParam {
			t.Errorf("zero host: got %v, want ErrInvalidParam", err)
		}
		// The #812 branch: the ATTACHMENT has gone. Reached by every
		// caller, not just equip — Armored Skyhunter's "you may attach
		// it to a creature you control" asks a prompt the Equipment
		// can leave before the answer arrives.
		if err := g.SacrificePermanentForEffect(sword); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear}); err != nil {
			t.Errorf("departed attachment: got %v, want nil", err)
		}
	})

	if got, ok := battlefieldCardByID(g, sword); ok && got.IsAttached() {
		t.Errorf("sword attached to %+v after four refused attaches", got.AttachedTo)
	}
	var onBear []uuid.UUID
	g.ReadSnapshot(func() { onBear = g.AttachmentsOf(bear) })
	if len(onBear) != 0 {
		t.Errorf("the bear picked up %v", onBear)
	}
	if n := countEventsOfKind(g, EventAttachSkipped); n != 3 {
		t.Errorf("EventAttachSkipped count = %d, want 3 (self, absent host, departed attachment)", n)
	}
	if n := countEventsOfKind(g, EventAttach); n != 0 {
		t.Errorf("EventAttach count = %d, want 0", n)
	}
	if n := countEventsOfKind(g, EventEffectError); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0 — a refused attach is not a failure", n)
	}
}

// countEventsOfKind is the event-log read the CR 701.3b tests share.
func countEventsOfKind(g *Game, kind EventKind) int {
	n := 0
	g.ReadSnapshot(func() {
		for i := range g.Events {
			if g.Events[i].Kind == kind {
				n++
			}
		}
	})
	return n
}

// Equip is re-activatable (CR 702.6d) and the second activation
// simply moves the Equipment — including a refreshed CR 613.7d
// timestamp.
func TestAttachForEffectReattachOverwrites(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	ox := pushAttachTestCard(g, me.ID, "Ox", "Creature — Ox")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")

	var first int64
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear})
	})
	if c, _ := battlefieldCardByID(g, sword); true {
		first = c.AttachedAt
	}
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: ox})
	})

	got, _ := battlefieldCardByID(g, sword)
	if !got.IsAttachedTo(ox) {
		t.Fatalf("re-equip did not move the sword: %+v", got.AttachedTo)
	}
	if got.AttachedAt < first {
		t.Errorf("AttachedAt went backwards on re-equip: %d then %d", first, got.AttachedAt)
	}
	var rev []uuid.UUID
	g.ReadSnapshot(func() { rev = g.AttachmentsOf(bear) })
	if len(rev) != 0 {
		t.Errorf("old host still lists %v", rev)
	}
}

func TestUnattachForEffectIsIdempotent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")

	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear})
		if err := g.UnattachForEffect(sword); err != nil {
			t.Fatalf("UnattachForEffect: %v", err)
		}
		// Unattaching something already unattached is a no-op, so a
		// catalog effect can call it without a guard.
		if err := g.UnattachForEffect(sword); err != nil {
			t.Fatalf("second UnattachForEffect: %v", err)
		}
	})
	got, ok := battlefieldCardByID(g, sword)
	if !ok {
		t.Fatal("unattaching moved the sword off the battlefield")
	}
	if got.IsAttached() || got.AttachedAt != 0 {
		t.Errorf("still attached: %+v @%d", got.AttachedTo, got.AttachedAt)
	}
}

// CR 704.5n — the Equipment branch. The host leaves; the Equipment
// unattaches and STAYS on the battlefield.
func TestSBAUnattachesEquipmentWhenHostLeaves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")

	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear})
		if _, err := MoveCard(g.Battlefield, me.Graveyard, bear); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: bear, OldZone: ZoneBattlefield, NewZone: ZoneGraveyard})
		g.runStateChecksLocked()
	})

	got, ok := battlefieldCardByID(g, sword)
	if !ok {
		t.Fatal("CR 704.5n: the Equipment must STAY on the battlefield")
	}
	if got.IsAttached() {
		t.Errorf("still attached to a card that left: %+v", got.AttachedTo)
	}
}

// CR 704.5n again, the other way in: the host is still on the
// battlefield but has stopped being a creature. Read through
// Effective(), never the printed type line.
func TestSBAUnattachesEquipmentFromNonCreature(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	rock := pushAttachTestCard(g, me.ID, "Mana Rock", "Artifact")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")

	g.WithWriteLock(func() {
		// The primitive does not enforce CR 301.5c — an effect that
		// says "attach" to a non-creature performs the attach, and
		// the SBA tears it down immediately afterwards.
		if err := g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: rock}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
		g.runStateChecksLocked()
	})

	got, _ := battlefieldCardByID(g, sword)
	if got.IsAttached() {
		t.Errorf("Equipment stayed attached to a non-creature: %+v", got.AttachedTo)
	}
	if _, ok := battlefieldCardByID(g, rock); !ok {
		t.Error("the host should be untouched")
	}
}

// CR 704.5m — the Aura branch, which is the opposite outcome from
// the same trigger condition.
func TestSBAPutsIllegallyAttachedAuraInGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	aura := pushAttachTestCard(g, me.ID, "Test Aura", "Enchantment — Aura")

	g.WithWriteLock(func() {
		_ = g.AttachForEffect(aura, TargetRef{Kind: TargetCard, ID: bear})
		if _, err := MoveCard(g.Battlefield, me.Graveyard, bear); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: bear, OldZone: ZoneBattlefield, NewZone: ZoneGraveyard})
		g.runStateChecksLocked()
	})

	if _, ok := battlefieldCardByID(g, aura); ok {
		t.Fatal("CR 704.5m: the Aura must leave the battlefield")
	}
	if !graveyardHas(me, aura) {
		t.Error("the Aura should be in its owner's graveyard")
	}
	var sawUnattach bool
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventUnattach && ev.CardID == aura {
				sawUnattach = true
			}
		}
	})
	if !sawUnattach {
		t.Error("no EventUnattach emitted for the falling Aura")
	}
}

// A Curse attached to a player stays put while the player is in the
// game, and falls off when they are eliminated.
func TestSBAKeepsPlayerAttachmentUntilElimination(t *testing.T) {
	// Four seats, not two: eliminating one of two ends the game, and
	// stateBasedActionsLocked is a no-op once State != StateActive.
	g := newFourPlayerActiveGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	curse := pushAttachTestCard(g, me.ID, "Test Curse", "Enchantment — Aura Curse")

	g.WithWriteLock(func() {
		if err := g.AttachForEffect(curse, TargetRef{Kind: TargetPlayer, ID: foe.ID}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
		g.runStateChecksLocked()
	})
	got, ok := battlefieldCardByID(g, curse)
	if !ok || !got.IsAttached() {
		t.Fatalf("Curse fell off a live player (onBattlefield=%v)", ok)
	}

	g.WithWriteLock(func() {
		foe.Life = 0
		g.runStateChecksLocked()
	})
	if _, ok := battlefieldCardByID(g, curse); ok {
		t.Error("CR 704.5m: a Curse on an eliminated player must go to the graveyard")
	}
}

// CR 400.7 / ADR 0036 decision 12 — the FORWARD sweep. MoveCard is
// the only place that clears the relation eagerly.
func TestMoveCardClearsAttachmentOnBattlefieldExit(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")

	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear})
		moved, err := MoveCard(g.Battlefield, me.Graveyard, sword)
		if err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
		if moved.IsAttached() || moved.AttachedAt != 0 {
			t.Errorf("attachment survived battlefield exit: %+v @%d", moved.AttachedTo, moved.AttachedAt)
		}
	})
}

// ADR 0036 decision 1 says a VALUE-typed AttachedTo is carried
// through cloneCard's `out := c` for free and that nothing
// structurally enforces this. This is the enforcement.
func TestCloneAndRestoreCarryAttachment(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear})
	})

	snap := g.Clone()
	var cloned Card
	for _, c := range snap.Battlefield.Cards {
		if c.InstanceID == sword {
			cloned = c
		}
	}
	if !cloned.IsAttachedTo(bear) || cloned.AttachedAt == 0 {
		t.Fatalf("clone lost the attachment: %+v @%d", cloned.AttachedTo, cloned.AttachedAt)
	}

	// Undo is Clone + RestoreFrom, so the round trip is what a
	// player actually exercises.
	g.WithWriteLock(func() { _ = g.UnattachForEffect(sword) })
	if got, _ := battlefieldCardByID(g, sword); got.IsAttached() {
		t.Fatal("setup: unattach did not take")
	}
	g.RestoreFrom(snap)
	if got, _ := battlefieldCardByID(g, sword); !got.IsAttachedTo(bear) {
		t.Errorf("undo did not restore the attachment: %+v", got.AttachedTo)
	}
}

// CR 613.7d: the continuous effect of an attached permanent is
// timestamped when it became attached, not when it entered.
func TestAttachedAtWinsTheLayerTimestamp(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	sword := pushAttachTestCard(g, me.ID, "Test Sword", "Artifact — Equipment")
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, TargetRef{Kind: TargetCard, ID: bear})
	})

	withStaticAbilities(t, func(oracleID string) []StaticAbility { return nil })
	var swordCard *Card
	var ts int64
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == sword {
				swordCard = &g.Battlefield.Cards[i]
			}
		}
		eff := staticContinuousEffect{source: swordCard, timestamp: swordCard.AttachedAt}
		ts = eff.Timestamp()
	})
	if swordCard == nil {
		t.Fatal("sword vanished")
	}
	if ts == swordCard.EnteredBattlefieldAt {
		t.Error("attachment timestamp should differ from the entry stamp")
	}
}

// The aura attach stamp at resolution (ADR 0036 decision 5), driven
// straight at the helper the resolution path calls. Keyed on the
// card TYPE, so an uncatalogued Aura attaches too.
func TestResolvedAuraAttachesToItsTarget(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	aura := pushAttachTestCard(g, me.ID, "Test Aura", "Enchantment — Aura")
	plain := pushAttachTestCard(g, me.ID, "Test Enchantment", "Enchantment")

	item := &StackItem{
		Controller: me.ID,
		Targets:    []TargetRef{{Kind: TargetCard, ID: bear}},
	}
	g.WithWriteLock(func() {
		g.attachResolvedAuraLocked(aura, item)
		// A non-Aura enchantment with a target is not attached to
		// anything — Beast Within targets and does not enchant.
		g.attachResolvedAuraLocked(plain, item)
	})

	got, _ := battlefieldCardByID(g, aura)
	if !got.IsAttachedTo(bear) {
		t.Errorf("Aura did not attach to its target: %+v", got.AttachedTo)
	}
	other, _ := battlefieldCardByID(g, plain)
	if other.IsAttached() {
		t.Errorf("non-Aura enchantment attached itself: %+v", other.AttachedTo)
	}
}

// Multi-slot and no-slot items are both skipped rather than guessed
// at — the sandbox's free-form picker leaves an Aura with no target.
func TestResolvedAuraIgnoresAmbiguousTargets(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a := pushAttachTestCard(g, me.ID, "A", "Creature — Bear")
	b := pushAttachTestCard(g, me.ID, "B", "Creature — Bear")
	aura := pushAttachTestCard(g, me.ID, "Test Aura", "Enchantment — Aura")

	g.WithWriteLock(func() {
		g.attachResolvedAuraLocked(aura, &StackItem{Controller: me.ID})
		g.attachResolvedAuraLocked(aura, &StackItem{
			Controller: me.ID,
			Targets:    []TargetRef{{Kind: TargetCard, ID: a}, {Kind: TargetCard, ID: b}},
		})
	})
	got, _ := battlefieldCardByID(g, aura)
	if got.IsAttached() {
		t.Errorf("attached from an ambiguous item: %+v", got.AttachedTo)
	}
}

// pushCataloguedAura seeds an Aura the catalog knows an enchant
// clause for — the distinction CR 704.5m's "attached to nothing"
// branch turns on, since an UNCATALOGUED Aura is a manual sandbox
// object and must not be swept.
func pushCataloguedAura(g *Game, controller uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Enchantment — Aura",
		OracleID:   name,
		Owner:      controller,
		Controller: controller,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  id,
			OldZone: ZoneGraveyard,
			NewZone: ZoneBattlefield,
		})
	})
	return id
}

// enchantCreatureSpec is the "enchant creature" clause every Aura in
// this file declares.
func enchantCreatureSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "creature",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
}

// CR 704.5m, second disjunct: "...or is not attached to an object or
// player". An Aura put onto the battlefield by an effect that does
// not say "attached to" — Brilliant Restoration, Carmen — used to sit
// there permanently, which is a board state no sequence of legal
// plays can reach.
func TestSBAPutsAnUnattachedAuraInTheGraveyard(t *testing.T) {
	withCatalogTargetSpec(t, func(string) *TargetSpec { return enchantCreatureSpec() })
	g := newActiveGame(t)
	me := g.Seats[0]
	_ = pushAttachTestCard(g, me.ID, "Grizzly Bears", "Creature — Bear")
	aura := pushCataloguedAura(g, me.ID, "Rancor")

	g.WithWriteLock(func() { g.runStateChecksLocked() })

	if _, ok := battlefieldCardByID(g, aura); ok {
		t.Fatal("CR 704.5m: an Aura attached to nothing must leave the battlefield")
	}
	if !graveyardHas(me, aura) {
		t.Error("the Aura should be in its owner's graveyard")
	}
	// There was no link to break, so there is nothing to announce.
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventUnattach && ev.CardID == aura {
				t.Error("EventUnattach emitted for an Aura that was never attached")
			}
		}
	})
}

// The sandbox posture, and the reason the branch above is scoped
// rather than universal: an Aura the catalog does not know is a
// manual object. It is cast through the free-form picker with no
// target, its controller is tracking it by hand, and sweeping it into
// a graveyard would delete a card the table is using.
func TestSBALeavesAnUncataloguedUnattachedAuraAlone(t *testing.T) {
	withCatalogTargetSpec(t, func(string) *TargetSpec { return nil })
	g := newActiveGame(t)
	me := g.Seats[0]
	aura := pushCataloguedAura(g, me.ID, "Homebrew Aura")

	g.WithWriteLock(func() { g.runStateChecksLocked() })

	if _, ok := battlefieldCardByID(g, aura); !ok {
		t.Error("an uncatalogued Aura with no host must stay on the battlefield")
	}
}

// CR 704.5n has no "attached to nothing" clause — an Equipment that
// is attached to nothing is an ordinary artifact sitting on the
// battlefield, which is where every Equipment starts its life.
func TestSBALeavesAnUnattachedEquipmentAlone(t *testing.T) {
	withCatalogTargetSpec(t, func(string) *TargetSpec { return enchantCreatureSpec() })
	g := newActiveGame(t)
	me := g.Seats[0]
	sword := pushAttachTestCard(g, me.ID, "Bonesplitter", "Artifact — Equipment")

	g.WithWriteLock(func() { g.runStateChecksLocked() })

	if _, ok := battlefieldCardByID(g, sword); !ok {
		t.Error("an unequipped Equipment must stay on the battlefield")
	}
}
