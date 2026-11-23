package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bns44Aura casts the Aura on `host` and lets it resolve.
func Bns44Aura(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := CastCatalogSpell(t, g, name, AuraTypeLine, oracle, CardRefs(host))
	PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatalf("%s did not resolve", name)
	}
	return aura
}
