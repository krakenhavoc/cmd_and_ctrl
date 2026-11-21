package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_mana_each_cards_test.go — ADR 0200, #3952: Drain Power and Pygmy
// Hippo, where a player activates a mana ability of each land they control
// and loses all unspent mana.

const (
	LmDrainPower   = "0669172d-396b-4f5a-9703-129c5c849b55"
	LmPygmyHippo   = "0d39513a-1f05-4d9a-a772-745f4f1a3471"
	LmCityOfBrass  = "f25351e3-539b-4bbc-b92d-6480acf4d722"
	LmGravenCairns = "5004b84a-33b7-4f6f-b2c2-7086b9087535"
)

func LmBasic(g *game.Game, owner uuid.UUID, kind string, tapped bool) uuid.UUID {
	id := ApaPush(g, owner, owner, game.Card{Name: kind, TypeLine: "Basic Land — " + kind})
	if tapped {
		FindBattlefieldCardForTest(g, id).Tapped = true
	}
	return id
}

func LmTapped(g *game.Game, id uuid.UUID) bool {
	c := FindBattlefieldCardForTest(g, id)
	return c != nil && c.Tapped
}

// lmOptionIndex is the index of the option whose label ends with `suffix`
// on the seat's open option pick.
func LmOptionIndex(t *testing.T, g *game.Game, chooser uuid.UUID, suffix string) int {
	t.Helper()
	c := LatestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for the land: %+v", g.PendingChoices)
	}
	var labels []string
	for i, o := range c.PickOptions {
		if strings.HasSuffix(o.Label, suffix) {
			return i
		}
		labels = append(labels, o.Label)
	}
	t.Fatalf("no option ending %q in %v", suffix, labels)
	return -1
}
