package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mlTop puts a card on top of p's library and returns its id.
func MlTop(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 4, Toughness: 4, Owner: p.ID, Controller: p.ID})
	return id
}

// mlManifested asserts id is a manifested 2/2 that `controller` controls,
// `owner` owns, and only `controller` may look at (CR 708.5).
func MlManifested(t *testing.T, g *game.Game, id uuid.UUID, controller, owner *game.Player) {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatal("the card is not on the battlefield")
	}
	if c.FaceDownKind != game.FaceDownManifested || !c.FaceDownIsPermanent() {
		t.Fatalf("face-down kind %q, want manifested", c.FaceDownKind)
	}
	if c.Controller != controller.ID || c.Owner != owner.ID {
		t.Fatalf("controller is the manifesting player: %v, owner is the library's: %v", c.Controller == controller.ID, c.Owner == owner.ID)
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !c.IsCreature() {
		t.Errorf("a manifested permanent is a 2/2 creature, got %d/%d creature=%v", c.CurrentPower(), c.CurrentToughness(), c.IsCreature())
	}
	if !c.IsKnownTo(controller.ID) {
		t.Error("the controller cannot look at their own face-down permanent")
	}
	for _, p := range g.Seats {
		if p.ID != controller.ID && c.IsKnownTo(p.ID) {
			t.Errorf("seat %s can look at a face-down permanent another player controls (CR 708.5)", p.Name)
		}
	}
}
