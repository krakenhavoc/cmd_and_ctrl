package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// retarget_cards_test.go — #1196's card proofs. One board shape
// serves all four: a Lightning Bolt aimed at one seat, a redirect
// spell cast on top of it, and the question of where the 3 damage
// lands.

const (
	DeflectingSwatOracle = "ae120613-97d6-4393-b39d-c3e6c076f5d6"
	BoltBendOracle       = "c20a96f7-aa5a-4c15-b8b1-806685c99b27"
	MisdirectionOracle   = "c39e5fb0-6de3-4105-ad3c-0ecb8951a1d5"
	ImpsMischiefOracle   = "30ec73ad-c7a1-4527-8dc6-44bdd65b9ba6"
)

func LatestRetarget(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceRetarget && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
