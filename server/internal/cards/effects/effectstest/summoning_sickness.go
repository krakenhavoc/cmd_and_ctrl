package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// summoning_sickness_test.go — #530, and the two player reports it
// explains (#365 Treasures, #368 Fabled Passage).
//
// CR 302.6: "A creature's activated ability with the tap symbol or
// the untap symbol in its activation cost can't be activated unless
// the creature has been under its controller's control continuously
// since their most recent turn began. A creature can't attack unless
// it has been under its controller's control continuously since
// their most recent turn began."
//
// The rule names a CREATURE. A Treasure token, a Sol Ring and a
// fetchland are none of them creatures, and nothing stops their
// controller tapping them the turn they arrive. The engine's write
// paths had that right; game.HasSummoningSickness did not, and the
// wire read the helper, so the client greyed every one of these.
//
// These tests run the engine rather than the helper — the helper's
// own table is in server/internal/game/keywords_test.go and the wire
// contract is in server/internal/protocol/summoning_sick_view_test.go.

// makeSick marks a permanent already on the battlefield as having
// entered this turn, the way layer_listener.go stamps every entry.
// Called AFTER the step machine has been advanced, since crossing an
// untap step clears the flag for that seat.
func MakeSick(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].SummonedThisTurn = true
			return
		}
	}
	t.Fatalf("permanent %s is not on the battlefield", id)
}
