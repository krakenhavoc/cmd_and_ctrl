package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_choice_keyed_restore_test.go — #4291: the "you may" a resolving
// effect asks (MayChoice) and a ward's life, sacrifice and discard
// questions are keyed, so a table waiting on one is a restore point. Each
// test takes a STRICT restore (restoreRoundTrip with strict, which fails on
// a table that is not a restore point) while the prompt is open, checks the
// restored prompt still names its key and the data it carries, then answers
// it on the restored table. The *UnknownKey tests are the rollback case: a
// binary that does not know the key refuses the file.

// restoreAtConfirm restores the table strictly while a confirm owed to
// `chooser` is open, and returns the restored game and the restored prompt,
// having checked the prompt names `key`.
func RestoreAtConfirm(t *testing.T, g *game.Game, chooser uuid.UUID, key string) (*game.Game, *game.PendingChoice) {
	t.Helper()
	ask := LatestConfirmFor(g, chooser)
	if ask == nil || ask.PickThen != key {
		t.Fatalf("before the restore: %+v, want an open confirm keyed %q", ask, key)
	}
	g = RestoreRoundTrip(t, g, true)
	ask = LatestConfirmFor(g, chooser)
	if ask == nil || ask.PickThen != key {
		t.Fatalf("after the restore: %+v, want an open confirm keyed %q", ask, key)
	}
	return g, ask
}

func SeatOf(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p.ID == id {
			return p
		}
	}
	return nil
}
