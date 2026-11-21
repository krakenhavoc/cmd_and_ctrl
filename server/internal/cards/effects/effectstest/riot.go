package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answerRiotFor answers the open entry_riot prompt, failing when there
// is none.
func AnswerRiotFor(t *testing.T, g *game.Game, counter bool) {
	t.Helper()
	c := PendingOfKind(g, game.PendingChoiceEntryRiot)
	if c == nil {
		t.Fatal("no riot question is open")
	}
	if err := g.ResolveEntryRiot(c.ID, c.Chooser, counter); err != nil {
		t.Fatalf("ResolveEntryRiot: %v", err)
	}
}

func RiotPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, counters int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	}
	if counters > 0 {
		c.Counters = map[string]int{game.CounterPlusOne: counters}
	}
	return PushBattlefieldCardWithTimestamp(g, c)
}
