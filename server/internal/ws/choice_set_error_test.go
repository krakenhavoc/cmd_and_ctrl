package ws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestClassifyActionErrorSendsAChoiceSetRejectionAsARule — #624. When a
// choose-cards prompt's set-level rule refuses an answer, the error
// frame's message is the only explanation the player's open prompt
// shows, and the prompt stays open for another try. So it must arrive
// as a player-facing sentence, as bad_request (the player can fix it),
// and must not collapse into the generic "invalid parameter" a
// malformed payload earns.
//
// The sentence is the classifier's, not the sentinel's: game sentinels
// are short internal strings (as ErrUnparseableCost and ErrInvalidFace
// are), and the wire wording is chosen here. docs/protocol.md quotes
// the same sentence.
func TestClassifyActionErrorSendsAChoiceSetRejectionAsARule(t *testing.T) {
	const want = "that selection doesn't meet the card's condition — check its text and choose again"
	code, msg := classifyActionError(game.ErrChoiceSetRejected)
	if code != protocol.CodeBadRequest {
		t.Errorf("code = %q, want %q", code, protocol.CodeBadRequest)
	}
	if msg != want {
		t.Errorf("message = %q, want %q", msg, want)
	}
	// Wrapped, as an effect that annotates the error would send it.
	if _, wrapped := classifyActionError(fmt.Errorf("resolve: %w", game.ErrChoiceSetRejected)); wrapped != want {
		t.Errorf("wrapped message = %q, want %q", wrapped, want)
	}
	if raw := strings.TrimPrefix(game.ErrChoiceSetRejected.Error(), "game: "); msg == raw {
		t.Errorf("message is the sentinel's internal text %q; the classifier should supply the wording", raw)
	}
	_, generic := classifyActionError(game.ErrInvalidParam)
	if msg == generic {
		t.Errorf("a rule rejection reads the same as a malformed payload: %q", msg)
	}
}
