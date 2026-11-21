package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr2_cards_test.go — ADR 0108 Delivery PR 2 (#1890): damage
// doubled or tripled by a resolved spell or ability, each card against
// the board that tells its wording apart.

const (
	P2InsultInjuryOracle  = "47543892-4d60-4c6b-a6a4-69b9172af01e"
	P2IsengardOracle      = "2b947703-751a-4d95-b5de-e2d6b1fcb502"
	P2GoblinGoliathOracle = "131069a6-8f30-4caf-8934-3588837b5f7f"
	P2BlindFuryOracle     = "5dbfd316-a0a9-4caa-99b1-069931d2aaa4"
	P2DesperateOracle     = "d8328d27-e27f-4c84-8ab5-cabb80a34f54"
	P2ManeuversOracle     = "39a9323d-dddc-42ac-929d-3f4fa7c87567"
	P2QuestOracle         = "71220cc2-5f3d-4c97-ae66-05b1b79adef1"
	P2LightningOracle     = "585eb5bc-5a3d-44d8-b593-1ff0d67f96a7"
)

// p2Ping is a source `from` dealing n non-combat damage to `to`.
func P2Ping(t *testing.T, g *game.Game, from, to uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{ID: uuid.New(), SourceCardID: from})
		if err := (effects.DealDamage{Source: from, Target: to, Amount: n}).Apply(ctx); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
}

// p2Tough is a vanilla creature too big to die of a doubled hit.
func P2Tough(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear",
		Power: 1, Toughness: 30, Owner: owner, Controller: owner,
	})
}
