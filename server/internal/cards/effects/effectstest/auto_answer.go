package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// autoAnswerAll answers every prompt the table's standing answers
// cover, the way the room does after a commit, and returns how many.
func AutoAnswerAll(t *testing.T, g *game.Game) int {
	t.Helper()
	n := 0
	for ; n < 64; n++ {
		id, _, ok := g.NextAutoAnswer()
		if !ok {
			return n
		}
		if _, err := g.AutoAnswer(id); err != nil {
			t.Fatalf("AutoAnswer: %v", err)
		}
	}
	t.Fatal("more than 64 automatic answers in a row")
	return n
}
