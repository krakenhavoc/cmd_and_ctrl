package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- MayPayX ---------------------------------------------------------

// payXPrompt is the keyed "choose X" `p` owes: a pay_amount that holds
// no frame, naming the branches the payment runs.
func PayXPrompt(t *testing.T, g *game.Game, p uuid.UUID) *game.PendingChoice {
	t.Helper()
	c := PayAmountPrompt(g, p)
	if c == nil {
		t.Fatalf("no number prompt for %s", p)
	}
	if c.MayPayX == nil || c.PickThen == "" || c.PayAmount == nil {
		t.Fatalf("the number prompt is not keyed: %+v", c)
	}
	return c
}
