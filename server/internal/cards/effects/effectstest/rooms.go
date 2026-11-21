package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms_test.go — the card-facing half of ADR 0103: the Room
// constructor stamps a door gate on every ability, Register refuses an
// ungated one, and the trigger shapes fire off the engine's events —
// "when you unlock this door" as the Room enters with the cast door
// (CR 709.5h) and after an unlock, "whenever you fully unlock a Room"
// (CR 709.5i), and the eerie pair.

const testRoomOracle = "test-room-fixture-oracle"

func TestRoomCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   testRoomOracle,
		Owner:      owner,
		Controller: owner,
		Layout:     game.LayoutSplit,
		Faces: []game.Face{
			{Name: "Test Foyer", TypeLine: "Enchantment — Room", ManaCost: "{1}{W}", Colors: []string{"W"}},
			{Name: "Test Vault", TypeLine: "Enchantment — Room", ManaCost: "{3}{W}", Colors: []string{"W"}},
		},
	}
	c.SettleImported()
	return c
}

func RegisterTestRoom(t *testing.T, counts map[string]int) {
	t.Helper()
	bump := func(key string) effects.Effect {
		return func(*game.Game, *game.StackItem) error {
			counts[key]++
			return nil
		}
	}
	RegisterForTest(t, effects.Room(effects.RoomSpec{
		OracleID:     testRoomOracle,
		Name:         "Test Foyer // Test Vault",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{
			effects.WhenYouUnlockThisDoor(game.DoorLeft, "Test Foyer — left unlocked", bump("left")),
		}},
		Right: effects.Door{
			Triggered: []game.TriggeredAbility{
				effects.WhenYouUnlockThisDoor(game.DoorRight, "Test Vault — right unlocked", bump("right")),
				effects.WheneverYouFullyUnlockARoom("Test Vault — fully unlocked", bump("full")),
			},
			NoMaxHandSize: true,
		},
	}))
}

// settleOrdering resolves everything, answering each CR 603.3b
// trigger_order prompt `chooser` gets in the order offered — a door
// trigger and a full-unlock trigger (or an eerie trigger) go on the
// stack together.
func SettleOrdering(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 64; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
				if err := g.ResolveTriggerOrder(c.ID, chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
				break
			}
		}
		if answered {
			continue
		}
		if StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}
