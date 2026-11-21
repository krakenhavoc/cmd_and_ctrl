package testsav

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const beastModeOracle = "a465f8a9-d5aa-4dfa-b674-7aa1798cb0bd"

func TestBeastModeWithTeamworkAlsoPlacesACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := effectstest.PushVanillaCreature(g, me.ID, "Elf", 1, 1)
	target := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := effectstest.CastPaying(t, g, "Beast Mode", "Instant", beastModeOracle, effectstest.TwTarget(target),
		[]int{0}, []uuid.UUID{elf}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var power int
	g.ReadSnapshot(func() { power = effectstest.TwCard(g, target).CurrentPower() })
	if power != 5 {
		t.Errorf("power (pump + counter): %d, want 5 (2 base + 2 pump + 1 counter)", power)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, target), "trample") {
		t.Errorf("trample not granted")
	}
	if got := effectstest.TwCard(g, target).Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("+1/+1 counters: %d, want 1", got)
	}
}

func TestBeastModeWithoutTeamworkPumpsButNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := effectstest.CastPaying(t, g, "Beast Mode", "Instant", beastModeOracle, effectstest.TwTarget(target),
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, target); got != 4 {
		t.Errorf("power: %d, want 4 (2 base + 2 pump, no counter)", got)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, target), "trample") {
		t.Errorf("trample not granted")
	}
	if got := effectstest.TwCard(g, target).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("+1/+1 counters: %d, want 0 when teamwork wasn't used", got)
	}
}

func TestBeastModeRefusesTeamworkWithTooLittlePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := effectstest.CastPaying(t, g, "Beast Mode", "Instant", beastModeOracle, effectstest.TwTarget(target),
		[]int{0}, nil, nil); !errors.Is(err, game.ErrInsufficientTeamwork) {
		t.Fatalf("err = %v, want ErrInsufficientTeamwork", err)
	}
}
