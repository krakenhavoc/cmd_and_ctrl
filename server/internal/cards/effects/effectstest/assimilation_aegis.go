package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aegisCardNow is the card `id` names wherever it is, as it is now.
func AegisCardNow(g *game.Game, id uuid.UUID) game.Card {
	var c game.Card
	g.ReadSnapshot(func() { c, _ = g.LookupCardForEffect(id) })
	return c
}

func AegisWantName(t *testing.T, g *game.Game, id uuid.UUID, want string, power int) {
	t.Helper()
	if got := AegisCardNow(g, id).Name; got != want {
		t.Fatalf("creature is %q, want %q", got, want)
	}
	if p := EffectivePowerWithCounters(t, g, id); p != power {
		t.Errorf("%s's power %d, want %d", want, p, power)
	}
}
