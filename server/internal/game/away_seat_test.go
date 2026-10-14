package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// away_seat_test.go pins ADR 0177's engine half: the away time setting,
// a concession made for an away player, and the log lines.

func TestAwayAfterDefaultsToThreeMinutesAndZeroMeansNever(t *testing.T) {
	g := NewGame()
	if got := g.TableSettingsSnapshot().EffectiveAwayAfter(); got != DefaultAwayAfter {
		t.Fatalf("a new table's away time is %d, want %d", got, DefaultAwayAfter)
	}
	// A snapshot written before the setting existed decodes the stored
	// zero, and must keep the default rather than turn the offer off.
	if got := (TableSettings{}).EffectiveAwayAfter(); got != DefaultAwayAfter {
		t.Fatalf("an unset away time reads %d, want the default %d", got, DefaultAwayAfter)
	}
	never := 0
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{AwayAfter: &never}); err != nil {
		t.Fatal(err)
	}
	if got := g.TableSettingsSnapshot().EffectiveAwayAfter(); got != 0 {
		t.Fatalf("away_after 0 reads %d, want 0 (never)", got)
	}
	five := 5
	if err := g.UpdateSettings(uuid.Nil, SettingsPatch{AwayAfter: &five}); err != nil {
		t.Fatal(err)
	}
	if got := g.TableSettingsSnapshot().EffectiveAwayAfter(); got != 5 {
		t.Fatalf("away_after 5 reads %d", got)
	}
	var changes []string
	for _, ev := range g.Events {
		if ev.Kind == EventSettingsChanged && ev.Label == SettingAwayAfter {
			changes = append(changes, ev.SettingOld+"->"+ev.SettingNew)
		}
	}
	if len(changes) != 2 || changes[0] != "3->0" || changes[1] != "0->5" {
		t.Errorf("away_after changes logged %v, want [3->0 0->5]", changes)
	}
}

func TestAwayAfterOutOfRangeIsRefused(t *testing.T) {
	g := NewGame()
	for _, v := range []int{-1, MaxAwayAfter + 1} {
		if err := g.UpdateSettings(uuid.Nil, SettingsPatch{AwayAfter: &v}); !errors.Is(err, ErrInvalidSetting) {
			t.Errorf("away_after %d: err %v, want ErrInvalidSetting", v, err)
		}
	}
}

// TestAwayAfterChangesMidGame: §7, changeable at any time.
func TestAwayAfterChangesMidGame(t *testing.T) {
	p := newDepartureProbe(t)
	one := 1
	if err := p.g.UpdateSettings(uuid.Nil, SettingsPatch{AwayAfter: &one}); err != nil {
		t.Fatalf("an active game refused an away time change: %v", err)
	}
}

func TestConcedeForAwayIsAConcessionNotARemoval(t *testing.T) {
	p := newDepartureProbe(t)
	actor := p.c.ID
	if err := p.g.ConcedeForAway(actor, p.b.ID, 4); err != nil {
		t.Fatalf("ConcedeForAway: %v", err)
	}
	p.assertBHasLeftTheGame(t)
	if !p.b.Eliminated || p.b.Removed {
		t.Errorf("B: eliminated=%v removed=%v, want eliminated and not removed", p.b.Eliminated, p.b.Removed)
	}
	var away, concedes int
	for _, ev := range p.g.Events {
		switch {
		case ev.Kind == EventAwaySeat:
			away++
			if ev.Actor != actor || ev.Target != p.b.ID || ev.Label != AwayConcede || ev.Amount != 4 {
				t.Errorf("away event %+v; want actor C, target B, concede, 4 minutes", ev)
			}
		case ev.Kind == EventConcede && ev.Actor == p.b.ID:
			concedes++
		}
	}
	if away != 1 || concedes != 1 {
		t.Errorf("away events %d, concessions %d; want 1 and 1", away, concedes)
	}
	if err := p.g.ConcedeForAway(actor, p.b.ID, 4); !errors.Is(err, ErrPlayerEliminated) {
		t.Errorf("a second concession: err %v, want ErrPlayerEliminated", err)
	}
}

func TestNoteAwaySeatLogsWithoutChangingTheGame(t *testing.T) {
	p := newDepartureProbe(t)
	before := len(p.g.Events)
	if err := p.g.NoteAwaySeat(uuid.Nil, p.b.ID, AwayStandIn, 3); err != nil {
		t.Fatal(err)
	}
	if p.b.Eliminated {
		t.Fatal("a hand-over eliminated the seat")
	}
	if got := len(p.g.Events) - before; got != 1 {
		t.Fatalf("a hand-over emitted %d events, want 1", got)
	}
	if ev := p.g.Events[len(p.g.Events)-1]; ev.Kind != EventAwaySeat || ev.Label != AwayStandIn || ev.Target != p.b.ID {
		t.Errorf("event %+v", ev)
	}
	if err := p.g.NoteAwaySeat(uuid.Nil, uuid.New(), AwayStandIn, 3); !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("an unknown seat: err %v, want ErrPlayerNotFound", err)
	}
}
