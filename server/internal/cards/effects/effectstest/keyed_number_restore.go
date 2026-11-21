package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_number_restore_test.go — #4294: the number-answer prompts (a
// pay_amount) carry their consequence as a registered key, the asking item
// as data and, for a chain of questions or a split, the place in it, so a
// table waiting on one is a restore point. Each test takes a STRICT
// restore while the prompt is open and answers it on the restored table.

// numberPrompt is the open number prompt `chooser` owes, failing the test
// when there is none or when it holds a closure (it carries no key).
func NumberPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	c := PayAmountPrompt(g, chooser)
	if c == nil {
		t.Fatalf("no number prompt for %s", chooser)
	}
	if c.PickThen == "" {
		t.Fatalf("the number prompt holds no key: it is a closure prompt")
	}
	return c
}
