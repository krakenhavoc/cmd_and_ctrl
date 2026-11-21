package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// echo_test.go — ADR 0108 §5 (#1888), CR 702.30a: "At the beginning of
// your upkeep, if this permanent came under your control since the
// beginning of your last upkeep, sacrifice it unless you pay [cost]."

const (
	OracleGoblinPatrol      = "edf25cbe-fa28-44ca-b4be-1d2312e4f6ac"
	OracleDeepcavernImp     = "1f295f2f-969a-4b82-98d1-7fe307ec83a7"
	OracleSkizzikSurger     = "ded349de-0599-4744-ad3e-aec95caf9f99"
	OraclePolarKraken       = "d0ff30c8-ddb7-439d-b14d-c83fe3c8da87"
	OraclePhyrexianSoulgorg = "b6e36e77-0cea-4b7a-b968-774b1e74a992"
)

// echoPrompt is the open pay-unless prompt `chooser` owes, or nil.
func EchoPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// castEchoCreature casts a catalogued creature for the active seat and
// resolves it, returning the permanent.
func CastEchoCreature(t *testing.T, g *game.Game, name, typeLine, oracleID string) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, name, typeLine, oracleID, nil)
	PassPriorityAroundTable(t, g)
	if _, ok := BattlefieldCard(g, id); !ok {
		t.Fatalf("%s did not reach the battlefield", name)
	}
	return id
}

// leaveUpkeepThenAdvanceTo walks off the step the cursor is on and on to
// `seat`'s next upkeep.
func LeaveUpkeepThenAdvanceTo(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	AdvanceToUpkeepOf(t, g, seat)
}
