package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zone_layers_cards_test.go — #3862: the three cards whose types or
// colours reach objects off the battlefield (game/zone_layers.go). The
// oracle constants are lattice_lord_test.go's and tribal_test.go's.

// zoneLayerCard reads an object wherever it is, after a recompute.
func ZoneLayerCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.ReadSnapshot(func() {
		out, found = g.LookupCardForEffect(id)
	})
	if !found {
		t.Fatalf("object %s is in no zone", id)
	}
	return out
}

func RedCard(owner uuid.UUID, name, typeLine string) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: "{1}{R}",
		Owner: owner, Controller: owner,
	}
}
