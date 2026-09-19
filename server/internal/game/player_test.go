package game

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewPlayerStartsAt40Life(t *testing.T) {
	p := newPlayer("Alice", 0)
	if p.Life != 40 {
		t.Errorf("starting life: got %d, want 40", p.Life)
	}
	if p.Seat != 0 {
		t.Errorf("seat: got %d, want 0", p.Seat)
	}
	if p.Library == nil || p.Hand == nil || p.Graveyard == nil || p.Command == nil {
		t.Error("player zones should be pre-constructed")
	}
	if p.Library.Size() != 0 {
		t.Errorf("new library should be empty, got %d cards", p.Library.Size())
	}
}

func TestChangeLife(t *testing.T) {
	p := newPlayer("Bob", 1)
	if got := p.ChangeLife(-5); got != 35 {
		t.Errorf("after -5: got %d, want 35", got)
	}
	if got := p.ChangeLife(+3); got != 38 {
		t.Errorf("after +3: got %d, want 38", got)
	}
	// Life can go negative; death-by-life is a state-based action
	// evaluated in S13+ rules work, not here.
	if got := p.ChangeLife(-50); got != -12 {
		t.Errorf("negative life allowed: got %d, want -12", got)
	}
}

func TestRecordCommanderDamage(t *testing.T) {
	p := newPlayer("Carol", 2)
	commander := uuid.New()

	if got := p.RecordCommanderDamage(commander, 7); got != 7 {
		t.Errorf("first damage: got %d, want 7", got)
	}
	if got := p.RecordCommanderDamage(commander, 8); got != 15 {
		t.Errorf("cumulative: got %d, want 15", got)
	}
	if p.IsDeadByCommanderDamage(CommanderDamageLethal) {
		t.Error("15 damage should not be lethal")
	}

	p.RecordCommanderDamage(commander, 6)
	if !p.IsDeadByCommanderDamage(CommanderDamageLethal) {
		t.Error("21 damage from one commander should be lethal")
	}
}

// TestCommanderDamageFromMultipleCommandersNotCombined is CR 903.14a
// and, since the S25 (#77) rekey from player IDs to commander
// instance IDs, it finally covers the case it was named for: two
// PARTNER commanders sharing one seat are two independent clocks.
// Under the old player-keyed map they were one, and a partner pair
// at 15 apiece killed its victim ten damage early.
func TestCommanderDamageFromMultipleCommandersNotCombined(t *testing.T) {
	p := newPlayer("Dave", 3)
	a, b := uuid.New(), uuid.New()

	p.RecordCommanderDamage(a, 15)
	p.RecordCommanderDamage(b, 15)
	if p.IsDeadByCommanderDamage(CommanderDamageLethal) {
		t.Error("15+15 from different commanders should not be lethal (21 is per-commander)")
	}

	p.RecordCommanderDamage(a, 6)
	if !p.IsDeadByCommanderDamage(CommanderDamageLethal) {
		t.Error("21 from a single commander should be lethal even split across hits")
	}
}

func TestRecordCommanderDamageClampsNegative(t *testing.T) {
	p := newPlayer("Eve", 0)
	p.RecordCommanderDamage(uuid.New(), -5)
	for _, d := range p.CommanderDamage {
		if d < 0 {
			t.Errorf("negative damage persisted: %d", d)
		}
	}
}
