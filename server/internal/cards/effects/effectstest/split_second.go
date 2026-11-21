package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// split_second_test.go — #1519, CR 702.61. The six proof cards, each
// cast from its catalog entry with no sandbox flag: the keyword the
// entry declares is what turns split second on.

const (
	SuddenShockOracle    = "7b139231-dcd2-4b03-bcf2-fcb040617b69"
	SuddenDeathOracle    = "a3475ca6-d88e-4855-a554-1338624d1635"
	SuddenEdictOracle    = "b95f704d-96b9-437e-8c49-aba874139e12"
	SuddenSpoilingOracle = "dce202c7-fe8e-462a-858e-7a5a69bd5b6b"
)

// requireSplitSecondOn fails unless the spell is on the stack stamped
// with split second and the table-wide cache agrees.
func RequireSplitSecondOn(t *testing.T, g *game.Game, spell uuid.UUID) {
	t.Helper()
	item := g.StackMeta[spell]
	if item == nil {
		t.Fatal("the spell is not on the stack")
	}
	if !item.SplitSecond || !g.SplitSecondActive {
		t.Fatalf("split second is off (item=%v, table=%v): the catalog declaration was not read",
			item.SplitSecond, g.SplitSecondActive)
	}
}
