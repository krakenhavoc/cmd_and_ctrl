package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nghc_cards_test.go — #4092 slice c: the Captain N'ghathrod deck's
// Forgotten Creation, Endless Evil, Nihilith and Overcharged Amalgam
// (and the exploit keyword they share). Helpers carry the nghc prefix
// so the sibling slices' test files cannot collide with them.

const (
	NghcForgottenCreation  = "c26964c7-4259-4281-b450-f0c2ed34b37c"
	NghcEndlessEvil        = "05157e8e-4c8c-455e-a5a1-de8271951c51"
	NghcNihilith           = "c235fda2-1a8d-4ac7-9a23-784529a520f0"
	NghcOverchargedAmalgam = "d4838305-5392-42b2-81d1-9b9da71b20b1"
	NghcLightningBolt      = "4457ed35-7c10-48c8-9776-456485fdf070"
)

// nghcPushCreature seeds a catalog creature on the battlefield.
func NghcPushCreature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// nghcPassOnce passes priority around the table until the top of the
// stack changes (or the stack empties), resolving exactly one object.
func NghcPassOnce(t *testing.T, g *game.Game) {
	t.Helper()
	before := g.Stack.Size() + len(g.StackMeta)
	for i := 0; i < 8 && g.Stack.Size()+len(g.StackMeta) >= before && !StackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}
