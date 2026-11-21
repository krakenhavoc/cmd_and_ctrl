package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// census_s13_restore_test.go — #4361 S13, owner decision D7: a cast or an
// activation parked before anything is paid is written into the snapshot
// (game/parked_announcement.go). Each test parks a real catalog card's
// announcement, restores the table STRICTLY (the boot path) while the
// prompt is open, answers it on the restored table and checks the
// announcement completed there:
//
//   - the CR 903.9b question a cost asks a commander's owner: Ninja of
//     the Deep Hours' ninjutsu returning an attacking commander (an
//     activation), and Daze's alternative cost returning a commander
//     Island (a cast);
//   - "of an opponent's choice": Nova Pentacle's controller choosing the
//     opponent and that opponent's pick (an activation, both prompts
//     restored), and Evangelize's sole opponent's pick (a cast).

// s13StrictRestore checks the open prompt was written under `marker` and
// restores the table strictly through JSON.
func S13StrictRestore(t *testing.T, g *game.Game, marker string) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a table waiting on a parked announcement is not a restore point: %+v", snap.Continuations)
	}
	found := false
	for _, c := range snap.PendingChoices {
		if c.PickThen == marker {
			found = true
		}
	}
	if !found {
		t.Fatalf("no prompt in the snapshot is marked %q", marker)
	}
	return RestoreRoundTrip(t, g, true)
}
