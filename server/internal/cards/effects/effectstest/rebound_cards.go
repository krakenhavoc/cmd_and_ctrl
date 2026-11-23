package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castFromHandAtMain seeds a catalog card in the active seat's hand at
// its main phase and casts it.
func CastFromHandAtMain(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	AdvanceTo(t, g, game.StepPrecombatMain)
	id := HandCardForTest(me, name, typeLine, oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return me, id
}
