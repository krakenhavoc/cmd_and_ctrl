package ws

import (
	"fmt"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #4131: a refused table verb reaches the client as a bad request with a
// sentence a player can read, not the generic internal-error fallback
// that would leave the control failing silently.
func TestTableVerbRefusalsAreReadableBadRequests(t *testing.T) {
	for err, want := range map[error]string{
		actions.ErrNotTurnOrPriority: "only the active player or the priority holder can do that",
		actions.ErrNotVoteInitiator:  "only the player who called the vote can end it",
		actions.ErrMarkerNotYours:    "you can only take this marker for yourself, or release it while you hold it",
	} {
		code, msg := classifyActionError(fmt.Errorf("dispatch: %w", err))
		if code != protocol.CodeBadRequest || msg != want {
			t.Errorf("%v: classifyActionError = (%q, %q), want (bad request, %q)", err, code, msg, want)
		}
	}
}
