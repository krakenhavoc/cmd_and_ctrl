package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// villainPrompt is the one open option pick, which must be `chooser`'s.
func VillainPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	var open []*game.PendingChoice
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionPick {
				open = append(open, c)
			}
		}
	})
	if len(open) != 1 {
		t.Fatalf("%d option picks open, want exactly the one for %s: %+v", len(open), chooser, open)
	}
	if open[0].Chooser != chooser {
		t.Fatalf("the question is %s's, want %s's", open[0].Chooser, chooser)
	}
	return open[0]
}

func VillainAnswer(t *testing.T, g *game.Game, c *game.PendingChoice, index int) {
	t.Helper()
	if err := g.ResolveOptionPick(c.ID, c.Chooser, index); err != nil {
		t.Fatalf("ResolveOptionPick(%d): %v", index, err)
	}
}

func VillainFillLibrary(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		PushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Basic Land — Forest"})
	}
}

// battlefieldCountNamed counts the battlefield permanents a player
// controls with this name.
func BattlefieldCountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}
