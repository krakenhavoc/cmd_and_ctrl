package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prepare_cards_test.go — the four ADR 0090 proof cards (#1328): one
// test per card, through the catalog and the real cast path. The
// engine's own lifecycle tests (the sweep, phasing, snapshots) are
// game/prepare_test.go.

// preparationCard builds a preparation card the way deck.toGameCard
// does: per-face data, front-face keywords at card level, SetFace(0).
func PreparationCard(owner uuid.UUID, oracle string, keywords []string, front, spell game.Face) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Layout:     game.LayoutPrepare,
		Owner:      owner,
		Controller: owner,
		Keywords:   keywords,
		Faces:      []game.Face{front, spell},
	}
	c.SetFace(0)
	return c
}

// prepareCopyOf finds the CR 722.3c copy in exile by name.
func PrepareCopyOf(g *game.Game, name string) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.PrepareCopy && c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

func CountPrepareCopies(g *game.Game) int {
	n := 0
	for _, c := range g.Exile.Cards {
		if c.PrepareCopy {
			n++
		}
	}
	return n
}

// castFromHandAndResolve casts a card from the active seat's hand and
// lets the table resolve it.
func CastFromHandAndResolve(t *testing.T, g *game.Game, p *game.Player, c game.Card) {
	t.Helper()
	p.Hand.PushTop(c)
	if err := g.CastSpell(p.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(c.InstanceID) {
		t.Fatalf("%s did not resolve onto the battlefield", c.Name)
	}
}
