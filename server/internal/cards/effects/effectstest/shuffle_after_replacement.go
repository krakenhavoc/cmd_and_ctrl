package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blightsteelShuffleCount counts the "shuffle" EventSearchLibrary
// events stamped for `player` in `events` — the same shape
// ShuffleLibraryForEffect emits (effect_api.go, #1335). Position
// alone (`Contains`) cannot tell a shuffle from a landing on top;
// this is the assertion that can.
func BlightsteelShuffleCount(events []game.Event, player uuid.UUID) int {
	n := 0
	for _, ev := range events {
		if ev.Kind == game.EventSearchLibrary && ev.Label == "shuffle" && ev.Actor == player {
			n++
		}
	}
	return n
}
