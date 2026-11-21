package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soCreature puts a creature with the given power and toughness onto
// the battlefield under `owner`.
func SoCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func SoPlayer(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

// soCopyTheSpell copies spell `id` for `controller` (CR 707.10) and
// returns the copy's stack item.
func SoCopyTheSpell(t *testing.T, g *game.Game, id, controller uuid.UUID) *game.StackItem {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for k := range g.StackMeta {
		before[k] = true
	}
	if err := g.CopySpellForEffect(id, controller, false, nil); err != nil {
		t.Fatalf("copy: %v", err)
	}
	for k, item := range g.StackMeta {
		if !before[k] && item.IsCopy {
			return item
		}
	}
	t.Fatal("no copy on the stack")
	return nil
}
