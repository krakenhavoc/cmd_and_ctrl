package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// orderWith puts first at the front of the prompt's IDs.
func OrderWith(order *game.PendingChoice, first game.ReplacementEffectID) []game.ReplacementEffectID {
	out := []game.ReplacementEffectID{first}
	for _, id := range order.ReplacementEffectIDs {
		if id != first {
			out = append(out, id)
		}
	}
	return out
}
