package aiseat_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// passClock records when the runner's first applied pass happened.
// Observe runs on the runner's goroutine, inline after the dispatch.
type passClock struct {
	mu sync.Mutex
	at time.Time
}

func (c *passClock) Observe(ev aiseat.DecisionEvent) {
	if !ev.Applied || ev.Index < 0 || ev.Index >= len(ev.Input.Moves) {
		return
	}
	if ev.Input.Moves[ev.Index].Kind != legal.KindPass {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.at.IsZero() {
		c.at = time.Now()
	}
}

func (c *passClock) passed() (time.Time, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.at, !c.at.IsZero()
}

// TestRunnerHoldsItsPassOnAnotherSeatsStackItem is ADR 0119 §2 end to
// end through a production-paced runner: at a normal-speed table, a
// bot holding priority over another seat's item does not pass until
// the item has been on the stack for the preset's 2 s. The runner
// first sees the item at its first step, which is after `started`, so
// the pass cannot land before started+2s; that lower bound is the
// whole assertion, and a loaded machine only makes it later.
func TestRunnerHoldsItsPassOnAnotherSeatsStackItem(t *testing.T) {
	room := newRoom(t, 2, 2204)
	g := room.Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	advanceUntilPriority(t, g)
	normal := game.BotPaceNormal
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{BotPace: &normal}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	// The active seat (played by the test) announces a trigger and
	// passes, so the other seat (the bot) holds priority over it.
	human := g.Seats[g.Turn.ActiveSeat]
	bot := g.Seats[1-g.Turn.ActiveSeat].ID
	if err := g.AnnounceTrigger(human.ID, human.Hand.Cards[0].InstanceID, game.AbilityParams{Label: "a test trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	if sp := g.StackPresenceSnapshot(); sp.TopController != human.ID {
		t.Fatalf("the human's trigger is not on top of the stack: %+v", sp)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if holder := g.Seats[g.Turn.PriorityHolder].ID; holder != bot {
		t.Fatalf("priority is with %s after the human passed, want the bot %s", holder, bot)
	}

	clock := &passClock{}
	cfg := aiseat.DefaultConfig()
	cfg.Narrate = false
	cfg.Observer = clock
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := time.Now()
	r := aiseat.Start(ctx, room, bot, &scripted{}, cfg, nil, testLogger())

	waitFor(t, "the bot passes on the human's trigger", func() bool {
		_, ok := clock.passed()
		return ok
	})
	at, _ := clock.passed()
	if held := at.Sub(started); held < 2*time.Second {
		t.Errorf("the bot passed %v after the item reached it; a normal-speed table holds 2 s (ADR 0119 §2)", held)
	}
	cancel()
	waitForRunner(t, "the runner exits", r)
}
