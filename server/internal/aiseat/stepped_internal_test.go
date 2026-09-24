package aiseat

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// stepped_internal_test.go pins the two promises stepped.go makes
// about the production stepping door (#1503). In-package because the
// first is about r.cfg, which is unexported.

// A stepped runner holds on no wall clock except the policy's own
// deadline. Handed production's DefaultConfig — 700ms MinThink, 4s
// BlockGrace, the table's pace followed — it must keep none of the
// three: each is a hold on behalf of a seat that cannot act until
// Step returns, so each would cost its duration and change nothing.
func TestNewSteppedDropsTheWallClockHolds(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MinThink == 0 || cfg.BlockGrace == 0 || !cfg.FollowTablePace {
		t.Fatalf("DefaultConfig no longer sets the holds this test strips (%+v); the test measures nothing", cfg)
	}
	r := NewStepped(paceRoom(t), uuid.New(), &RandomPolicy{}, cfg, nil)
	if r.cfg.MinThink != 0 || r.cfg.BlockGrace != 0 || r.cfg.FollowTablePace {
		t.Errorf("a stepped runner kept a wall-clock hold: MinThink=%s BlockGrace=%s FollowTablePace=%v",
			r.cfg.MinThink, r.cfg.BlockGrace, r.cfg.FollowTablePace)
	}
	if r.cfg.MaxThink != cfg.MaxThink {
		t.Errorf("MaxThink = %s, want the caller's %s: it is the policy's deadline, not pacing", r.cfg.MaxThink, cfg.MaxThink)
	}
	if !r.Stepped() {
		t.Error("NewStepped built a runner that does not report itself stepped")
	}
	select {
	case <-r.Done():
	default:
		t.Error("a stepped runner has no loop, so Done must already be closed")
	}
}

// The concurrent runner is untouched by the door: Start's runner is
// not stepped, and a caller that tries to step it anyway is refused
// loudly rather than allowed to race the runner's own goroutine for
// the seat.
func TestStepRefusesARunnerBuiltByStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := Start(ctx, paceRoom(t), uuid.New(), &RandomPolicy{}, Config{}, nil, nil)
	if r.Stepped() {
		t.Error("Start built a runner that reports itself stepped")
	}
	cancel()
	<-r.Done()

	defer func() {
		v := recover()
		msg, _ := v.(string)
		if !strings.Contains(msg, "built by Start") {
			t.Errorf("Step on a started runner: recovered %v, want the built-by-Start panic", v)
		}
	}()
	r.Step(context.Background())
}
