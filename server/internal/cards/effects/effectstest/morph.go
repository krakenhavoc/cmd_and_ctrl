package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// morph_test.go — #1194 / ADR 0082, the CARD side. The engine-side
// mechanism is pinned in server/internal/game/face_down_cast_test.go;
// what belongs here is that the three declarations reach it: the key
// the client will send, the {3} the keyword charges, and the face-up
// price the permanent is turned over for.

const (
	WillbenderOracle      = "0aae277e-e58e-4115-b5fd-0459451e17ec"
	AinokTrackerOracle    = "bf84a598-12d3-406d-8eeb-40592e782b87"
	AerieBowmastersOracle = "43ace4d5-9006-4bad-bd17-3d368c20564d"
)

// manaTokens is a short spelling of a handful of mana tokens.
func ManaTokens(colors ...string) []game.ManaToken {
	out := make([]game.ManaToken, 0, len(colors))
	for _, c := range colors {
		out = append(out, game.ManaToken{Color: c})
	}
	return out
}

// findCardAnywhere reads the live copy of a card wherever it is.
func FindCardAnywhere(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s is in no zone", id)
	}
	return c
}
