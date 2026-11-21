package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// text_word_helpers_test.go — #3872 slice 2: the words a card file hands
// a helper as an ARGUMENT, the words of a cost, and the words of a source
// that has left the battlefield all read as the object's text now says
// them (CR 612.2, ADR 0160's 2026-10-16 follow-up). One rules test per
// door (textWordDoors in text_word_guard_test.go) and per engine seam.

const (
	TwhBloodMoonOracle    = "94fac5fe-97d5-4c12-a80c-8efff9d853ae"
	TwhYavimayaOracle     = "8dd5f5af-d2d8-4356-8617-8381081b930c"
	TwhHeroismOracle      = "74716642-fc8c-4f62-a556-154ed0b4af0e"
	TwhKukemssaOracle     = "49e3ac82-7c22-4dec-b61a-b581148c0419"
	twhLilianaOracle      = "92301066-8904-41bc-84bd-4aa286cbab9c"
	TwhCabalCoffersOracle = "7358e164-5704-4e78-9b21-6a9bf2a968ce"
	TwhLeechriddenOracle  = "d83c86c1-126d-49e9-9b13-9e55784c49c5"
	TwhUtopiaSprawlOracle = "00d8efa6-a2d9-4249-8da7-b45173675329"
	TwhPollutedDelta      = "ef86989d-ce80-4e55-aece-7d11710eeffa"
)

// Magical Hack's indices: the five basic land types in CR 305.6 order,
// then the four left once the first answer is taken out.
const (
	twhPlains, TwhIsland, TwhSwamp, TwhMountain, TwhForest = 0, 1, 2, 3, 4
)

// twhNewType is the second prompt's index for `to` once `from` is gone.
func TwhNewType(from, to int) int {
	if to > from {
		return to - 1
	}
	return to
}

// twhHack resolves Magical Hack on `target`, replacing land type `from`
// with `to` (CR 305.6 indices).
func TwhHack(t *testing.T, g *game.Game, target uuid.UUID, from, to int) {
	t.Helper()
	TwChange(t, g, "Magical Hack", TwHackOracle, target, from, TwhNewType(from, to))
}
