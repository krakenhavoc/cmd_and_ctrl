package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// superfriends_batch2_test.go — S27's remaining six planeswalkers.
// The helpers (pushWalkerForTest, loyaltyOf, advanceToMainOf,
// countNamed) are superfriends_test.go's; this file is separate only
// so two concurrent card batches do not collide on one test file.

const (
	UginSpiritDragonOracle      = "eecb3047-a563-441a-9175-200421981ac3"
	LilianaOfTheVeilOracle      = "0ba134d8-ee7d-48ec-8dc6-57942b8e9261"
	TeferiTemporalPilgrimOracle = "5f0fa785-37ab-46a8-9ec0-6f8b29576d93"
	TeferiHeroOfDominariaOracle = "f2f165b6-ef0a-42ad-9352-ba68be8248b0"
	KarnLiberatedOracle         = "0ca233f4-1b7f-4807-ab6e-2b1f5439b3db"
	SaheeliRaiOracle            = "5f3fe679-aff1-41b3-8d75-c78c2c0636f0"
)

// resolveAnyScryFor answers an open scry prompt for the player by
// putting everything back on top, which is what a test that does not
// care about the order wants.
func ResolveAnyScryFor(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c.Kind != game.PendingChoiceScry || c.Chooser != player {
			continue
		}
		if err := g.ResolveScry(c.ID, player, nil, c.ScryCards); err != nil {
			t.Fatalf("ResolveScry: %v", err)
		}
		return
	}
	t.Fatal("no scry prompt was queued")
}

// advanceToStepOnThisTurn walks forward to the named step without
// letting the turn roll over — the delayed triggers under test fire
// at "the next end step", which is this turn's.
func AdvanceToStepOnThisTurn(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 40; i++ {
		if g.Turn.Step == step {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to %s: %v", step, err)
		}
	}
	t.Fatalf("never reached step %s (at %s)", step, g.Turn.Step)
}
