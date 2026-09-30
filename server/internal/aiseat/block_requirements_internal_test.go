package aiseat

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// block_requirements_internal_test.go — #1597. When a CR 509.1c block
// requirement is owed, the enumerator withholds the defending player's
// pass and offers the required blocks as one AlwaysLegal move. A policy
// that declines there takes that move rather than sleeping on the
// table — the runner rule #1571 wrote for attacks, reached by blocks.
func TestDeclineWithAnOwedBlockTakesIt(t *testing.T) {
	r := &Runner{policy: declining{}, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	moves := []legal.Move{
		{Kind: legal.KindBlock, Label: "Block Ogre with Bear"},
		{Kind: legal.KindBlock, Label: "Block as required: Ogre with Bear and Ogre with Elf", AlwaysLegal: true},
	}
	out := r.decide(context.Background(), Input{Moves: moves}, time.Second)
	if out.index != 1 {
		t.Fatalf("index %d, want the required block at 1 (reason %q)", out.index, out.reason)
	}
	if out.fallback != FallbackDeclineAlwaysLegal {
		t.Errorf("fallback %q, want %q", out.fallback, FallbackDeclineAlwaysLegal)
	}
}
