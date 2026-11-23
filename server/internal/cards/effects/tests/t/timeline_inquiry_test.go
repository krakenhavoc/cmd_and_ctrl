package testst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestTimelineInquiryWithTeamworkSkipsTheDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushVanillaCreature(g, me.ID, "Elf A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "Elf B", 1, 1)
	handBefore := me.Hand.Size()
	if _, err := effectstest.CastPaying(t, g, "Timeline Inquiry", "Instant", effectstest.TimelineInquiryOracle, nil,
		[]int{0}, []uuid.UUID{a, b}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Errorf("hand size: %d, want %d (cast a card off the top, drew three, discarded none)", got, handBefore+3)
	}
	if effectstest.DiscardOwed(g, me.ID) != 0 {
		t.Errorf("a discard was owed after casting using teamwork")
	}
}

func TestTimelineInquiryWithoutTeamworkDiscardsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	handBefore := me.Hand.Size()
	if _, err := effectstest.CastPaying(t, g, "Timeline Inquiry", "Instant", effectstest.TimelineInquiryOracle, nil,
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Fatalf("discard owed: %d, want 1", effectstest.DiscardOwed(g, me.ID))
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != handBefore+3-1 {
		t.Errorf("hand size: %d, want %d (cast a card off the top, drew three, discarded one)", got, handBefore+3-1)
	}
}

func TestTimelineInquiryRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	weak := effectstest.PushVanillaCreature(g, me.ID, "Elf", 1, 1)
	if _, err := effectstest.CastPaying(t, g, "Timeline Inquiry", "Instant", effectstest.TimelineInquiryOracle, nil,
		[]int{0}, []uuid.UUID{weak}, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
