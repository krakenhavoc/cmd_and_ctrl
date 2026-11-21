package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spell_copy_test.go — CR 707.10. The three facts the copy has to
// get right are all observable on one board: the copy runs the
// spell's effect, the copy's targets can differ from the
// original's, and the copy leaves NOTHING behind when it is done.
//
// The last one is the test that would have caught the obvious first
// draft. Routing a resolved copy through the ordinary instant exit
// is a one-word mistake and produces a graveyard with two Lightning
// Bolts in it — which nothing in the damage assertions would
// notice, and which Tarmogoyf, Regrowth and flashback all would.

const (
	ReverberateOracle        = "a1f55890-31c5-4ed4-a2cd-7a4a9f05f8ca"
	TwincastOracle           = "8f878efc-850f-43d2-a6fe-5ea8d1dd5afb"
	IncreasingVengeanceOracl = "a5ca7bd9-0964-405f-adb9-7c27153595e6"
)

func LifeOf(g *game.Game, playerID uuid.UUID) int {
	for _, p := range g.Seats {
		if p.ID == playerID {
			return p.Life
		}
	}
	return -1
}

// graveyardSize counts a seat's graveyard. A copy that "ceased to
// exist" correctly leaves this number exactly where a copy that
// went to the graveyard would not.
func GraveyardSize(g *game.Game, playerID uuid.UUID) int {
	for _, p := range g.Seats {
		if p.ID == playerID {
			return p.Graveyard.Size()
		}
	}
	return -1
}
