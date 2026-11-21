package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flashback_test.go — S29. The game package pins the mechanic
// (cast_zones_test.go, flashback_test.go there); this file pins the
// three cards end to end, against the REAL catalog rather than a
// stubbed hook. That is the half a hook stub cannot check: that the
// card file actually declared the zone, that Register accepted the
// pair, and that the offer the client would be shown is the one the
// cast path accepts.

// seedGraveyardCard puts a card in the active seat's graveyard at a
// main phase and returns its instance ID. The mirror of
// castCatalogSpell's hand seeding.
func SeedGraveyardCard(t *testing.T, g *game.Game, name, typeLine, oracleID string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id
}
