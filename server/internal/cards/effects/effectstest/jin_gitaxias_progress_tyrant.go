package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const JinGitaxiasOracle = "f5daadc1-98ff-480a-82bb-fe7bfaa7b60e"

// settleThroughCopyRetarget passes priority around the table,
// answering any "choose new target for the copy" prompt by re-picking
// the same target, until the stack is empty.
func SettleThroughCopyRetarget(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if p := LatestPickTarget(g, chooser); p != nil {
			PickPlayer(t, g, chooser, target)
			continue
		}
		if StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("stack never settled")
}
