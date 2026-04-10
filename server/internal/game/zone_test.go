package game

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
)

func newTestZone(t *testing.T, kind ZoneKind) *Zone {
	t.Helper()
	return newZone(kind, uuid.Nil)
}

func newTestCard(t *testing.T, name string) Card {
	t.Helper()
	return NewCard(name, uuid.New())
}

func TestZonePushTopAndTop(t *testing.T) {
	z := newTestZone(t, ZoneLibrary)
	a, b, c := newTestCard(t, "A"), newTestCard(t, "B"), newTestCard(t, "C")
	z.PushTop(a)
	z.PushTop(b)
	z.PushTop(c)

	if z.Size() != 3 {
		t.Errorf("size: got %d, want 3", z.Size())
	}
	top, err := z.Top()
	if err != nil {
		t.Fatalf("top: %v", err)
	}
	if top.Name != "C" {
		t.Errorf("top card: got %q, want C", top.Name)
	}
	bot, err := z.Bottom()
	if err != nil {
		t.Fatalf("bottom: %v", err)
	}
	if bot.Name != "A" {
		t.Errorf("bottom card: got %q, want A", bot.Name)
	}
}

func TestZonePushBottom(t *testing.T) {
	z := newTestZone(t, ZoneLibrary)
	a, b := newTestCard(t, "A"), newTestCard(t, "B")
	z.PushTop(a)
	z.PushBottom(b)

	bot, _ := z.Bottom()
	if bot.Name != "B" {
		t.Errorf("bottom: got %q, want B", bot.Name)
	}
	top, _ := z.Top()
	if top.Name != "A" {
		t.Errorf("top: got %q, want A", top.Name)
	}
}

func TestZonePopTop(t *testing.T) {
	z := newTestZone(t, ZoneLibrary)
	a, b := newTestCard(t, "A"), newTestCard(t, "B")
	z.PushTop(a)
	z.PushTop(b)

	popped, err := z.PopTop()
	if err != nil {
		t.Fatalf("pop: %v", err)
	}
	if popped.Name != "B" {
		t.Errorf("popped: got %q, want B", popped.Name)
	}
	if z.Size() != 1 {
		t.Errorf("size after pop: got %d, want 1", z.Size())
	}
}

func TestZoneEmptyErrors(t *testing.T) {
	z := newTestZone(t, ZoneLibrary)
	if _, err := z.Top(); err != ErrZoneEmpty {
		t.Errorf("empty Top: got %v, want ErrZoneEmpty", err)
	}
	if _, err := z.Bottom(); err != ErrZoneEmpty {
		t.Errorf("empty Bottom: got %v, want ErrZoneEmpty", err)
	}
	if _, err := z.PopTop(); err != ErrZoneEmpty {
		t.Errorf("empty PopTop: got %v, want ErrZoneEmpty", err)
	}
}

func TestZoneRemoveById(t *testing.T) {
	z := newTestZone(t, ZoneHand)
	a, b, c := newTestCard(t, "A"), newTestCard(t, "B"), newTestCard(t, "C")
	z.PushTop(a)
	z.PushTop(b)
	z.PushTop(c)

	removed, err := z.Remove(b.InstanceID)
	if err != nil {
		t.Fatalf("remove: %v", err)
	}
	if removed.Name != "B" {
		t.Errorf("removed card: got %q, want B", removed.Name)
	}
	if z.Size() != 2 {
		t.Errorf("size after remove: got %d, want 2", z.Size())
	}
	if z.Contains(b.InstanceID) {
		t.Error("zone still contains removed card")
	}
	// Order of remaining cards must be preserved: [A, C]
	if z.Cards[0].Name != "A" || z.Cards[1].Name != "C" {
		t.Errorf("order after remove: got [%q, %q], want [A, C]", z.Cards[0].Name, z.Cards[1].Name)
	}
}

func TestZoneRemoveNotFound(t *testing.T) {
	z := newTestZone(t, ZoneHand)
	z.PushTop(newTestCard(t, "A"))
	if _, err := z.Remove(uuid.New()); err != ErrCardNotFound {
		t.Errorf("remove unknown: got %v, want ErrCardNotFound", err)
	}
}

func TestMoveCardBetweenZones(t *testing.T) {
	src := newTestZone(t, ZoneHand)
	dst := newTestZone(t, ZoneBattlefield)
	card := newTestCard(t, "Sol Ring")
	src.PushTop(card)

	moved, err := MoveCard(src, dst, card.InstanceID)
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	if moved.InstanceID != card.InstanceID {
		t.Errorf("move returned wrong card")
	}
	if src.Size() != 0 {
		t.Errorf("src size after move: got %d, want 0", src.Size())
	}
	if dst.Size() != 1 {
		t.Errorf("dst size after move: got %d, want 1", dst.Size())
	}
}

func TestMoveCardClearsBattlefieldState(t *testing.T) {
	// When a card leaves the battlefield, its tapped state and counters
	// must be cleared — it's a new "object" per CR 400.7.
	bf := newTestZone(t, ZoneBattlefield)
	gy := newTestZone(t, ZoneGraveyard)
	card := newTestCard(t, "Llanowar Elves")
	card.Tapped = true
	card.Counters = map[string]int{"+1/+1": 2}
	bf.PushTop(card)

	moved, err := MoveCard(bf, gy, card.InstanceID)
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	if moved.Tapped {
		t.Error("moved card is still tapped")
	}
	if moved.Counters != nil {
		t.Errorf("moved card still has counters: %v", moved.Counters)
	}
}

func TestMoveCardPreservesHandState(t *testing.T) {
	// Moving between non-battlefield zones should NOT clear state —
	// the tapped/counter reset is specific to leaving the battlefield.
	// (Hand cards can't be tapped in real rules, but for S02 we just
	// verify MoveCard doesn't touch state unless src is battlefield.)
	h := newTestZone(t, ZoneHand)
	g := newTestZone(t, ZoneGraveyard)
	card := newTestCard(t, "Lightning Bolt")
	card.Counters = map[string]int{"rebound": 1}
	h.PushTop(card)

	moved, _ := MoveCard(h, g, card.InstanceID)
	if moved.Counters["rebound"] != 1 {
		t.Errorf("rebound counter lost: %v", moved.Counters)
	}
}

func TestZoneShuffleDeterministic(t *testing.T) {
	// Using a seeded RNG, shuffle must be deterministic.
	z := newTestZone(t, ZoneLibrary)
	names := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	for _, n := range names {
		z.PushTop(newTestCard(t, n))
	}

	// Two shuffles with the same seed should produce the same order.
	r1 := rand.New(rand.NewPCG(42, 42))
	z1 := &Zone{Kind: z.Kind, Owner: z.Owner, Cards: append([]Card(nil), z.Cards...)}
	z1.Shuffle(r1)

	r2 := rand.New(rand.NewPCG(42, 42))
	z2 := &Zone{Kind: z.Kind, Owner: z.Owner, Cards: append([]Card(nil), z.Cards...)}
	z2.Shuffle(r2)

	for i := range z1.Cards {
		if z1.Cards[i].InstanceID != z2.Cards[i].InstanceID {
			t.Errorf("non-deterministic shuffle at index %d", i)
		}
	}
}

func TestZoneIsShared(t *testing.T) {
	shared := newZone(ZoneBattlefield, uuid.Nil)
	if !shared.IsShared() {
		t.Error("battlefield with nil owner should be shared")
	}
	owned := newZone(ZoneHand, uuid.New())
	if owned.IsShared() {
		t.Error("hand with owner should not be shared")
	}
}
