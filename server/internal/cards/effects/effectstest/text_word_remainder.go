package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// twxEvolve resolves Artificial Evolution on `target`, replacing the
// creature type `from` with `to`.
func TwxEvolve(t *testing.T, g *game.Game, target uuid.UUID, from, to string) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	CastCatalogSpell(t, g, "Artificial Evolution", "Instant", TwEvolutionOracle, TwTargetRef(target))
	for i := 0; i < 8 && PendingOfKind(g, game.PendingChoiceCreatureType) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	AnswerCreatureType(t, g, me.ID, from)
	AnswerCreatureType(t, g, me.ID, to)
}
