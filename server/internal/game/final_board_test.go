package game

import (
	"testing"

	"github.com/google/uuid"
)

// final_board_test.go — #4912: a player who leaves the game keeps a
// record of what they had, as plain data, although CR 800.4a takes every
// object they own out of the game.

func finalCardByID(cards []FinalCard, id uuid.UUID) (FinalCard, bool) {
	for _, c := range cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return FinalCard{}, false
}

// finalBoardProbe is the departure probe with the details the final board
// has to keep: counters and tapped state on B's creature, B's Aura on C's
// creature, and a face-down card B owns in exile.
func finalBoardProbe(t *testing.T) (departureProbe, uuid.UUID, uuid.UUID) {
	t.Helper()
	p := newDepartureProbe(t)
	perm := findCard(p.g, p.permanent)
	perm.Tapped = true
	perm.ScryfallID = "scry-b-creature"
	perm.Counters = map[string]int{CounterPlusOne: 2}

	aura := NewCard("B's Aura", p.b.ID)
	aura.TypeLine = "Enchantment — Aura"
	aura.AttachedTo = TargetRef{Kind: TargetCard, ID: p.cCreature}
	p.g.Battlefield.PushTop(aura)

	hidden := NewCard("B's Secret", p.b.ID)
	hidden.FaceDown = true
	p.g.Exile.PushTop(hidden)
	return p, aura.InstanceID, hidden.InstanceID
}

func TestLeavingTheGameRecordsTheFinalBoard(t *testing.T) {
	p, aura, hidden := finalBoardProbe(t)
	if p.b.FinalBoard != nil {
		t.Fatal("a player still in the game has a final board")
	}
	handBefore := p.b.Hand.Size()
	if err := p.g.Concede(p.b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	p.assertBHasLeftTheGame(t)
	fb := p.b.FinalBoard
	if fb == nil {
		t.Fatal("B left the game with no final board recorded")
	}

	creature, ok := finalCardByID(fb.Battlefield, p.permanent)
	if !ok {
		t.Fatalf("B's creature is not on the final battlefield: %+v", fb.Battlefield)
	}
	if !creature.Tapped || creature.Counters[CounterPlusOne] != 2 || creature.ScryfallID != "scry-b-creature" {
		t.Errorf("B's creature recorded as %+v, want tapped with two +1/+1 counters and its art", creature)
	}
	if creature.Power != 4 || creature.Toughness != 4 {
		t.Errorf("B's 2/2 with two +1/+1 counters recorded as %d/%d, want 4/4", creature.Power, creature.Toughness)
	}
	a, ok := finalCardByID(fb.Battlefield, aura)
	if !ok {
		t.Fatal("B's Aura is not on the final battlefield")
	}
	if a.AttachedTo != p.cCreature.String() || a.AttachedToName != "Kw Creature" {
		t.Errorf("B's Aura attached to %q (%q), want C's creature", a.AttachedTo, a.AttachedToName)
	}
	if _, ok := finalCardByID(fb.Battlefield, p.cCreature); ok {
		t.Error("C's creature is on B's final battlefield")
	}

	if _, ok := finalCardByID(fb.Hand, p.inHand); !ok || len(fb.Hand) != handBefore {
		t.Errorf("final hand has %d cards (B's seeded card found: %v), want all %d", len(fb.Hand), ok, handBefore)
	}
	if _, ok := finalCardByID(fb.Graveyard, p.inGrave); !ok {
		t.Errorf("final graveyard %+v is missing B's card", fb.Graveyard)
	}
	// The graveyard is B's zone, so C's card sitting in it is recorded
	// there too: it is what the seat showed.
	if _, ok := finalCardByID(fb.Graveyard, p.cSurvivor); !ok {
		t.Errorf("final graveyard %+v is missing the card that sat in it", fb.Graveyard)
	}
	cmdr, ok := finalCardByID(fb.Command, p.commander)
	if !ok || !cmdr.IsCommander || cmdr.Name != "B's Commander" {
		t.Errorf("final command zone %+v, want B's commander", fb.Command)
	}
	if _, ok := finalCardByID(fb.Exile, p.inExile); !ok {
		t.Errorf("final exile %+v is missing B's card", fb.Exile)
	}
	// The game goes on, so B's face-down exiled card is revealed as they
	// leave, and is recorded face up.
	secret, ok := finalCardByID(fb.Exile, hidden)
	if !ok || secret.FaceDown || secret.Name != "B's Secret" {
		t.Errorf("B's face-down exiled card recorded as %+v, want it revealed", secret)
	}
}

// TestFinalBoardNeverHoldsTheLibrary: the library's contents stay hidden.
func TestFinalBoardNeverHoldsTheLibrary(t *testing.T) {
	p, _, _ := finalBoardProbe(t)
	if err := p.g.Concede(p.b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	fb := p.b.FinalBoard
	for _, zone := range [][]FinalCard{fb.Battlefield, fb.Hand, fb.Graveyard, fb.Exile, fb.Command} {
		if _, ok := finalCardByID(zone, p.inLibrary); ok {
			t.Fatal("a library card is on the final board")
		}
	}
}

// TestFinalBoardWhenTheDepartureEndsTheGame: the last loser's board is
// recorded too, and a face-down exiled card nothing reveals keeps its
// identity out of the record.
func TestFinalBoardWhenTheDepartureEndsTheGame(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	a := g.Seats[0]
	inHand := seedCardIn(a.Hand, a.ID, "A in hand")
	hidden := NewCard("A's Secret", a.ID)
	hidden.FaceDown = true
	hidden.ScryfallID = "scry-secret"
	g.Exile.PushTop(hidden)
	if err := g.Concede(a.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if g.State == StateActive {
		t.Fatal("the game should be over")
	}
	fb := a.FinalBoard
	if fb == nil {
		t.Fatal("the last loser has no final board")
	}
	if _, ok := finalCardByID(fb.Hand, inHand); !ok {
		t.Errorf("final hand %+v is missing A's card", fb.Hand)
	}
	secret, ok := finalCardByID(fb.Exile, hidden.InstanceID)
	if !ok || !secret.FaceDown || secret.Name != "" || secret.ScryfallID != "" {
		t.Errorf("an unrevealed face-down card recorded as %+v, want face down with no identity", secret)
	}
}

// TestFinalBoardSurvivesASnapshotAndAClone: it is carried on the player.
func TestFinalBoardSurvivesASnapshotAndAClone(t *testing.T) {
	p, aura, _ := finalBoardProbe(t)
	if err := p.g.Concede(p.b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	r, err := p.g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	fb := r.Seats[1].FinalBoard
	if fb == nil {
		t.Fatal("the final board did not survive a snapshot")
	}
	if c, ok := finalCardByID(fb.Battlefield, p.permanent); !ok || c.Counters[CounterPlusOne] != 2 || !c.Tapped {
		t.Errorf("restored creature %+v", c)
	}
	if c, ok := finalCardByID(fb.Battlefield, aura); !ok || c.AttachedToName != "Kw Creature" {
		t.Errorf("restored Aura %+v", c)
	}

	cp := clonePlayer(p.b)
	cp.FinalBoard.Battlefield[0].Counters[CounterPlusOne] = 9
	if c, _ := finalCardByID(p.b.FinalBoard.Battlefield, cp.FinalBoard.Battlefield[0].InstanceID); c.Counters[CounterPlusOne] == 9 {
		t.Error("a clone shares the final board's counters with the original")
	}
}
