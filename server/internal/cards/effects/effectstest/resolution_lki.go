package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_lki_test.go is the card-level proof for #1379: an ability
// that reads a permanent AT RESOLUTION reads its last-known information
// once the permanent has left the battlefield (CR 608.2h). Each test
// removes the permanent while the trigger waits on the stack, which is
// exactly the window the harvest-time CR 603.10 store no longer covers.

// removeInResponse destroys (or exiles) a permanent with the trigger
// still waiting, and checks it is gone.
func RemoveInResponse(t *testing.T, g *game.Game, id uuid.UUID, exile bool) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		if exile {
			err = g.ExileCardForEffect(id)
		} else {
			err = g.DestroyPermanentForEffect(id)
		}
	})
	if err != nil {
		t.Fatalf("removing %v: %v", id, err)
	}
	if g.Battlefield.Contains(id) {
		t.Fatalf("setup: %v is still on the battlefield", id)
	}
}
