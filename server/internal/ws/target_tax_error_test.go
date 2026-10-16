package ws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// A cast refused because its targets add a life cost the caster can't pay
// (#3860, CR 601.2f, 119.4) reads as a sentence about the card, not as
// "invalid parameter": the engine's refusal wraps ErrCantPayTargetTax with
// the permanent and the life.
func TestClassifyActionErrorNamesWhatTargetingCosts(t *testing.T) {
	err := fmt.Errorf("%w: targeting Terror of the Peaks costs 3 more life", game.ErrCantPayTargetTax)
	code, msg := classifyActionError(err)
	if code != protocol.CodeBadRequest {
		t.Errorf("code = %q, want a bad request the player can fix", code)
	}
	if !strings.Contains(msg, "Terror of the Peaks") || !strings.Contains(msg, "3 more life") {
		t.Errorf("message %q should name the permanent and the life", msg)
	}
	if strings.Contains(msg, "game:") {
		t.Errorf("message %q leaks the package prefix", msg)
	}
}
