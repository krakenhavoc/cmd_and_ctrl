package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const heroicTeamworkOracle = "8025313d-048d-4a43-bc01-22398939a7e7"

func htTargets(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		out[i] = game.TargetRef{Kind: game.TargetCard, ID: id}
	}
	return out
}

func TestHeroicTeamworkWithTeamworkPumpsBothAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := pushVanillaCreature(g, me.ID, "Elf", 3, 3)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	handBefore := me.Hand.Size()
	if _, err := castPaying(t, g, "Heroic Teamwork", "Instant", heroicTeamworkOracle, htTargets(a, b),
		[]int{0}, []uuid.UUID{elf}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, a); got != 4 {
		t.Errorf("A power: %d, want 4", got)
	}
	if got := effectivePower(t, g, b); got != 4 {
		t.Errorf("B power: %d, want 4", got)
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand size: %d, want %d (drew a card for using teamwork)", got, handBefore+1)
	}
}

func TestHeroicTeamworkWithoutTeamworkNoDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	handBefore := me.Hand.Size()
	if _, err := castPaying(t, g, "Heroic Teamwork", "Instant", heroicTeamworkOracle, htTargets(a),
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, a); got != 4 {
		t.Errorf("A power: %d, want 4", got)
	}
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand size: %d, want %d (no draw without teamwork)", got, handBefore)
	}
}

func TestHeroicTeamworkRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	weak := pushVanillaCreature(g, me.ID, "Elf", 1, 1)
	if _, err := castPaying(t, g, "Heroic Teamwork", "Instant", heroicTeamworkOracle, htTargets(a),
		[]int{0}, []uuid.UUID{weak}, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
