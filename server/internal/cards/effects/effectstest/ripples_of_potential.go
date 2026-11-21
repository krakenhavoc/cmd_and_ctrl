package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ripples_of_potential_test.go — #3930. "Proliferate, then choose any
// number of permanents you control that had a counter put on them this
// way. Those permanents phase out."

const (
	ripplesOfPotentialOracle = "45bcebd0-5e93-4433-b92f-b4c161d04bd6"
	VorinclexRaiderOracle    = "5a3fdf5a-bff8-4896-b288-3f43f9a72d9b"
)

// ripplesChoosePrompt is the open choose-cards prompt Ripples asks, or nil.
func RipplesChoosePrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			return c
		}
	}
	return nil
}

func CastRipples(t *testing.T, g *game.Game) {
	t.Helper()
	CastCatalogSpell(t, g, "Ripples of Potential", "Instant", ripplesOfPotentialOracle, nil)
	PassPriorityAroundTable(t, g)
}

// TestProliferateReportsPermanentsAndPlayers: the engine's report names
// players as well as permanents.
//
// testProliferatedTail records what the keyed form told it (#4253: the
// permanents are the first Amount IDs, the players the rest).
var (
	TestProliferatedHeard []game.Proliferated
	TestProliferatedTail  = game.RegisterTail("test/4253-proliferated", func(_ *game.Game, o game.TailOutcome) error {
		n := min(max(o.Amount, 0), len(o.IDs))
		TestProliferatedHeard = append(TestProliferatedHeard, game.Proliferated{
			Permanents: append([]uuid.UUID(nil), o.IDs[:n]...),
			Players:    append([]uuid.UUID(nil), o.IDs[n:]...),
		})
		return nil
	})
)
