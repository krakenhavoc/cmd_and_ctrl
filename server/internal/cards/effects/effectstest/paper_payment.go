package effectstest

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func PpAsks(t *testing.T, err error) *game.PaperPaymentRequiredError {
	t.Helper()
	var ask *game.PaperPaymentRequiredError
	if !errors.As(err, &ask) {
		t.Fatalf("cast = %v, want a paper payment prompt", err)
	}
	return ask
}
