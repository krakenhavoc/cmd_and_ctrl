package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const timelineInquiryOracle = "74a478c8-7498-43a6-ae0b-3964c5815b3d"

func TestTimelineInquiryWithTeamworkSkipsTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "Elf A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Elf B", 1, 1)
	handBefore := me.Hand.Size()
	if _, err := castPaying(t, g, "Timeline Inquiry", "Instant", timelineInquiryOracle, nil,
		[]int{0}, []uuid.UUID{a, b}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Errorf("hand size: %d, want %d (cast a card off the top, drew three, discarded none)", got, handBefore+3)
	}
	if discardOwed(g, me.ID) != 0 {
		t.Errorf("a discard was owed after casting using teamwork")
	}
}

func TestTimelineInquiryWithoutTeamworkDiscardsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	handBefore := me.Hand.Size()
	if _, err := castPaying(t, g, "Timeline Inquiry", "Instant", timelineInquiryOracle, nil,
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if discardOwed(g, me.ID) != 1 {
		t.Fatalf("discard owed: %d, want 1", discardOwed(g, me.ID))
	}
	discardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != handBefore+3-1 {
		t.Errorf("hand size: %d, want %d (cast a card off the top, drew three, discarded one)", got, handBefore+3-1)
	}
}

func TestTimelineInquiryRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	weak := pushVanillaCreature(g, me.ID, "Elf", 1, 1)
	if _, err := castPaying(t, g, "Timeline Inquiry", "Instant", timelineInquiryOracle, nil,
		[]int{0}, []uuid.UUID{weak}, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
