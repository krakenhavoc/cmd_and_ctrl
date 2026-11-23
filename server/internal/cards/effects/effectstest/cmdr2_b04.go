package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2WaitForChooseCards passes priority until a choose_cards prompt opens.
func Cmdr2WaitForChooseCards(t *testing.T, g *game.Game, who uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 8; i++ {
		if p := LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, who); p != nil {
			return p
		}
		PassPriorityAroundTable(t, g)
	}
	t.Fatal("no choose_cards prompt opened")
	return nil
}
