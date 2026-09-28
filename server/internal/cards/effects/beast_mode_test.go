package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const beastModeOracle = "a465f8a9-d5aa-4dfa-b674-7aa1798cb0bd"

func TestBeastModeWithTeamworkAlsoPlacesACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := pushVanillaCreature(g, me.ID, "Elf", 1, 1)
	target := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := castPaying(t, g, "Beast Mode", "Instant", beastModeOracle, twTarget(target),
		[]int{0}, []uuid.UUID{elf}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	var power int
	g.ReadSnapshot(func() { power = twCard(g, target).CurrentPower() })
	if power != 5 {
		t.Errorf("power (pump + counter): %d, want 5 (2 base + 2 pump + 1 counter)", power)
	}
	if !containsString(effectiveAbilities(t, g, target), "trample") {
		t.Errorf("trample not granted")
	}
	if got := twCard(g, target).Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("+1/+1 counters: %d, want 1", got)
	}
}

func TestBeastModeWithoutTeamworkPumpsButNoCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := castPaying(t, g, "Beast Mode", "Instant", beastModeOracle, twTarget(target),
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, target); got != 4 {
		t.Errorf("power: %d, want 4 (2 base + 2 pump, no counter)", got)
	}
	if !containsString(effectiveAbilities(t, g, target), "trample") {
		t.Errorf("trample not granted")
	}
	if got := twCard(g, target).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("+1/+1 counters: %d, want 0 when teamwork wasn't used", got)
	}
}

func TestBeastModeRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := castPaying(t, g, "Beast Mode", "Instant", beastModeOracle, twTarget(target),
		[]int{0}, nil, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
