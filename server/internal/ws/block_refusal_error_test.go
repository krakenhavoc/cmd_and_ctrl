package ws

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestBlockRefusalsReachTheWireAsIllegalBlock — ADR 0045 addendum
// Decision 8 and test plan item 15. Every reason the engine can return
// arrives as `illegal_block` with a non-empty sentence, its token in
// `reason` and the blocker in `card_id`, whether the hub's structured
// path or the plain classifier sees it, and however deep it is wrapped.
func TestBlockRefusalsReachTheWireAsIllegalBlock(t *testing.T) {
	blocker, attacker, defender := uuid.New(), uuid.New(), uuid.New()
	for _, reason := range game.BlockReasons() {
		refusal := &game.BlockRefusedError{
			BlockRefusal: game.BlockRefusal{Reason: reason, Source: attacker},
			Blocker:      blocker, Attacker: attacker,
			BlockerName: "Grizzly Bears", AttackerName: "Cold-Eyed Selkie",
			Keyword: "islandwalk", Defender: defender, DefenderName: "P2",
			LandName: "Island", LandKind: "an Island",
		}
		for _, err := range []error{refusal, fmt.Errorf("dispatch: %w", refusal)} {
			body, ok := blockRefusalPayload(err, defender)
			if !ok {
				t.Fatalf("%s: not recognised as a block refusal", reason)
			}
			if body.Code != protocol.CodeIllegalBlock || body.Message == "" ||
				body.Reason != string(reason) || body.CardID != blocker.String() {
				t.Errorf("%s: payload = %+v", reason, body)
			}
			code, msg := classifyActionError(err)
			if code != protocol.CodeIllegalBlock || msg == "" {
				t.Errorf("%s: classifyActionError = (%q, %q)", reason, code, msg)
			}
		}
	}

	// The landwalk sentence is addressed to its reader.
	lw := &game.BlockRefusedError{
		BlockRefusal: game.BlockRefusal{Reason: game.BlockReasonLandwalk},
		AttackerName: "Cold-Eyed Selkie", Keyword: "islandwalk",
		Defender: defender, DefenderName: "P2", LandName: "Island", LandKind: "an Island",
	}
	if body, _ := blockRefusalPayload(lw, defender); body.Message != "Cold-Eyed Selkie has islandwalk, and you control an Island (Island)." {
		t.Errorf("defender reads %q", body.Message)
	}
	if body, _ := blockRefusalPayload(lw, uuid.New()); body.Message != "Cold-Eyed Selkie has islandwalk, and P2 controls an Island (Island)." {
		t.Errorf("another player reads %q", body.Message)
	}

	// A bare sentinel still gets the code, just no reason.
	if body, ok := blockRefusalPayload(game.ErrIllegalBlock, defender); !ok || body.Code != protocol.CodeIllegalBlock || body.Reason != "" || body.Message == "" {
		t.Errorf("bare ErrIllegalBlock: payload = %+v, ok = %v", body, ok)
	}
	// And nothing else is mistaken for one.
	if _, ok := blockRefusalPayload(game.ErrWrongStep, defender); ok {
		t.Error("ErrWrongStep was sent as illegal_block")
	}
}
