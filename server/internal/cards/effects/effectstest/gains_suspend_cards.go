package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gsSetLibrary replaces the player's library with `cards`, the first listed on top.
func GsSetLibrary(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		ids[i] = c.InstanceID
		p.Library.PushTop(c)
	}
	return ids
}

func GsForest() game.Card { return game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"} }

func GsExiled(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == id {
			return &g.Exile.Cards[i]
		}
	}
	return nil
}

// gsNextUpkeep walks the table to the named seat's NEXT upkeep and
// resolves what the step queued, stopping at a prompt.
func GsNextUpkeep(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	AdvanceToStepOf(t, g, seat, game.StepUpkeep)
	PassPriorityAroundTable(t, g)
}
