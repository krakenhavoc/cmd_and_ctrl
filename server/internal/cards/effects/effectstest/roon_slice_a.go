package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// roonaStolen puts a creature `owner` owns onto the battlefield under
// `controller`'s control.
func RoonaStolen(g *game.Game, owner, controller uuid.UUID, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: controller,
	})
}

// roonaOnly is the one battlefield card named `name`, or fails.
func RoonaOnly(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	var found []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			found = append(found, c)
		}
	}
	if len(found) != 1 {
		t.Fatalf("%d permanents named %q on the battlefield, want 1", len(found), name)
	}
	return found[0]
}

func RoonaCardRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}
