package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// giveRole attaches a Role of `kind` to `host`, controlled by
// `controller`, through the same primitive the cards use.
func GiveRole(t *testing.T, g *game.Game, kind effects.RoleKind, host, controller uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: controller}
		err = effects.CreateRoleToken{Role: kind, Host: host, Controller: controller}.Apply(effects.NewContext(g, item))
	})
	if err != nil {
		t.Fatalf("CreateRoleToken %s: %v", kind, err)
	}
	g.RunStateChecksForTest()
}

func HandCount(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.Hand.Size() })
	return n
}
