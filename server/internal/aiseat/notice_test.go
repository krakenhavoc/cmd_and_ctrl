package aiseat_test

import (
	"context"
	"sync"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// noticer declines every window and has one notice to give, the way an
// own-key seat does once its key fails (ADR 0182).
type noticer struct {
	mu     sync.Mutex
	notice string
	asked  int
}

func (p *noticer) Name() string { return "noticer" }

func (p *noticer) Decide(context.Context, aiseat.Input) (aiseat.Decision, error) {
	return aiseat.Decision{Index: aiseat.Decline}, nil
}

func (p *noticer) TakeNotice() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.asked++
	n := p.notice
	p.notice = ""
	return n
}

func (p *noticer) timesAsked() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.asked
}

// wrapper hides the noticer one level down, as tiers' noticePolicy does:
// the runner must find it through Unwrap.
type wrapper struct{ aiseat.Policy }

func (w wrapper) Unwrap() aiseat.Policy { return w.Policy }

func TestARunnerPostsItsPolicysNoticeOnceAsTheSeat(t *testing.T) {
	room := newRoom(t, 2, 7)
	bot := room.Game.Seats[0].ID
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bc := &chatBroadcaster{}
	inner := &noticer{notice: "I can't use my owner's model key any more."}
	r := startBot(t, ctx, room, bot, wrapper{inner}, aiseat.Config{}, bc, testLogger())

	// A line that has been posted stays posted: a monotone wait.
	waitFor(t, "the notice to be posted", func() bool { return bc.count() >= 1 })
	cancel()
	<-r.Done()

	lines := bc.lines(protocol.ChatKindBotNotice)
	if len(lines) != 1 {
		t.Fatalf("got %d chat lines, want exactly 1: %+v", len(lines), lines)
	}
	if lines[0].AuthorID != bot.String() || lines[0].Text != "I can't use my owner's model key any more." {
		t.Errorf("chat line = %+v", lines[0])
	}
	if inner.timesAsked() < 1 {
		t.Error("the runner never asked the policy for a notice")
	}
}
