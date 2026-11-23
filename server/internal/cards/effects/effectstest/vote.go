package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vote_test.go — ADR 0146, #2143: voting (CR 701.38) and the cards that
// ship with it.

const (
	pleaForPowerOracle     = "8b886623-5c96-48dd-a3d2-fd5865c58ff5"
	CouncilsJudgmentOracle = "acf4d685-a718-42f7-9a7a-dee069ae8db2"
	ExpropriateOracle      = "a69265b2-0e37-4d86-866e-e4a923233b4d"
	CoercivePortalOracle   = "eea4feef-2595-4ed2-92a0-11ccb3e7d40e"
	MagisterOfWorthOracle  = "0429d146-332f-4fb0-8273-c73679ca6d47"
	GaladrielElvenOracle   = "b97a11c2-7119-4399-9565-bf69520d3988"
	BragosRepOracle        = "8877b896-8458-4828-9ff6-9e0ecc180cf7"
	BallotBrokerOracle     = "e8df915f-f509-4c6c-9261-a5a16c3f06c5"
	TivitOracle            = "f4e0b557-4f80-4b7c-bfcc-6d4c468faf2e"
)

// ballotFor is the open ballot owed by voter, or nil.
func BallotFor(g *game.Game, voter uuid.UUID) *game.PendingChoice {
	c := LatestOptionPickFor(g, voter)
	if c == nil || c.CouncilVote == nil {
		return nil
	}
	return c
}

// voteFor casts voter's open ballot for the option labelled label.
func VoteFor(t *testing.T, g *game.Game, voter uuid.UUID, label string) {
	t.Helper()
	c := BallotFor(g, voter)
	if c == nil {
		t.Fatalf("no ballot open for %s: %+v", voter, g.PendingChoices)
	}
	for i, o := range c.PickOptions {
		if o.Label == label {
			if err := g.ResolveOptionPick(c.ID, voter, i); err != nil {
				t.Fatalf("vote %q: %v", label, err)
			}
			return
		}
	}
	t.Fatalf("ballot for %s offers no %q: %+v", voter, label, c.PickOptions)
}

// voteOrder is the table in turn order from the active player.
func VoteOrder(g *game.Game) []*game.Player {
	n := len(g.Seats)
	out := make([]*game.Player, n)
	for i := range out {
		out[i] = g.Seats[(g.Turn.ActiveSeat+i)%n]
	}
	return out
}

// castPlea casts Plea for Power for the active player and lets it
// start resolving.
func CastPlea(t *testing.T, g *game.Game) {
	t.Helper()
	CastCatalogSpell(t, g, "Plea for Power", "Sorcery", pleaForPowerOracle, nil)
	PassPriorityAroundTable(t, g)
}
