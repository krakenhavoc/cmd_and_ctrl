package aiseat

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// attack_requirements_internal_test.go — #1571. When a CR 508.1d attack
// requirement is owed, the enumerator withholds the active player's
// pass and marks the attacks that answer it AlwaysLegal. A policy that
// declines there holds priority with no pass to fall back on, so the
// runner turns the decline into the always-legal answer instead of
// sleeping on the table.

type declining struct{}

func (declining) Name() string { return "declining" }
func (declining) Decide(context.Context, Input) (Decision, error) {
	return Decision{Index: Decline, Reason: "nothing worth doing"}, nil
}

func TestDeclineWithNoPassTakesTheOwedAttack(t *testing.T) {
	r := &Runner{policy: declining{}, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	moves := []legal.Move{
		{Kind: legal.KindCast, Label: "Cast Shock"},
		{Kind: legal.KindAttack, Label: "Attack Alice with Goaded Bear", AlwaysLegal: true},
	}
	out := r.decide(context.Background(), Input{Moves: moves}, time.Second)
	if out.index != 1 {
		t.Fatalf("index %d, want the owed attack at 1 (reason %q)", out.index, out.reason)
	}
	if out.fallback != FallbackDeclineAlwaysLegal {
		t.Errorf("fallback %q, want %q", out.fallback, FallbackDeclineAlwaysLegal)
	}
}

func TestDeclineWithOnlyBlocksStillDeclines(t *testing.T) {
	// A defender with block moves and no priority: nothing is owed, so
	// the decline stands, as before #1571.
	r := &Runner{policy: declining{}, log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	moves := []legal.Move{{Kind: legal.KindBlock, Label: "Block Ogre with Bear"}}
	if out := r.decide(context.Background(), Input{Moves: moves}, time.Second); out.index != Decline {
		t.Fatalf("index %d, want Decline", out.index)
	}
}
