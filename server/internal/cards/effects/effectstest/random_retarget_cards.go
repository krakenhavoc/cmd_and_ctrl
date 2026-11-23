package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func SoleTarget(t *testing.T, item *game.StackItem) game.TargetRef {
	t.Helper()
	var out []game.TargetRef
	for _, r := range item.Targets {
		if r.Kind == game.TargetCard || r.Kind == game.TargetPlayer {
			out = append(out, r)
		}
	}
	if len(out) != 1 {
		t.Fatalf("want one target, got %+v", item.Targets)
	}
	return out[0]
}
