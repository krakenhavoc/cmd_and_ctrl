package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// look_at_top_test.go — "look at the top N cards of your library,
// then put them back in any order": the scry family's third member,
// and the cards built on it.
//
// The assertion that carries this file is the Sensei's Divining Top
// loop, which is the sprint's stated exit criterion: activate, see
// three cards nobody else can see, reorder them, confirm — then tap
// to draw the one you put on top and watch the Top itself land back
// on the library.

const (
	SenseisTopOracle    = "13575cf9-65c1-4861-b21e-eb2155e07766"
	PonderOracle        = "02090581-61aa-4348-ad57-451be8ee91c2"
	CrystalBallOracle   = "bd85fe4d-1d62-416f-ac2d-e287911c84e3"
	OtherworldlyOracle  = "be668c2d-71ea-4346-8980-1fbf5e4cbed3"
	ThoughtScourOracle  = "83101ba8-a569-4827-8c53-9ca0dfcd59a7"
	ReliquaryTowerOracl = "c23e5b80-08d2-4e24-9908-fe2aa4f30f6f"
)

func LookAtTopChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceLookAtTop && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func ConfirmChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
