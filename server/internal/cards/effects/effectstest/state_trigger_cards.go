package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// state_trigger_cards_test.go — the CR 603.8 state-trigger cards of ADR
// 0107 delivery PR 1 (#1858). Each test makes the printed state arise
// the way play would, passes priority so the trigger goes on the stack
// and resolves, and checks the result.

const (
	StBarbarianOutcast  = "49d44e89-b94e-43f7-a996-f50651df9264"
	StCovetousDragon    = "bf7710cb-70cb-42fb-b840-cc4f385daa7a"
	StEmperorCrocodile  = "26eb7ee3-c0b6-4be0-bc80-92952f55a6f9"
	StSynodCenturion    = "fe6337dc-df02-4baa-9fe0-b14e8d34186a"
	StTetheredGriffin   = "909e1bff-237a-4259-9c0f-419185681782"
	StEndangeredArmodon = "68879437-bbe2-4e99-b17b-ddec30bfd3d0"
	StSkeletonShip      = "8c85887d-5935-46ec-a075-dc9f5133bb96"
	StTaskMageAssembly  = "dfbbfbe1-d6c1-4b5b-8d09-949d70b9ff42"
	StDeadlyDesigns     = "293c6f8f-9e6b-4267-a15d-1d5ce95fb5b4"
	StDarksteelReactor  = "bc483bab-14fb-498d-9310-9c070766c7ae"
)

// stCard is a template for a test permanent.
func StCard(name, oracle, typeLine string, power, toughness int) game.Card {
	return game.Card{Name: name, OracleID: oracle, TypeLine: typeLine, Power: power, Toughness: toughness}
}

// stStateItems counts `source`'s triggered items, queued and on the
// stack, whose label starts with `prefix`.
func StStateItems(g *game.Game, source uuid.UUID, prefix string) int {
	n := 0
	for _, it := range g.PendingTriggers {
		if it.SourceCardID == source && strings.HasPrefix(it.Label, prefix) {
			n++
		}
	}
	for _, it := range g.StackMeta {
		if it.SourceCardID == source && it.Kind == game.StackItemTriggered && strings.HasPrefix(it.Label, prefix) {
			n++
		}
	}
	return n
}

// stDestroy destroys a permanent the way an effect would.
func StDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}
