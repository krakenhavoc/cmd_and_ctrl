package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// chain_cycle_test.go — #920. A spell copying ITSELF at resolution
// (CR 707.10), which the engine could not do until the resolving
// item's metadata outlived its own resolution function.
//
// The assertions that matter are the three the bug hid: a copy is
// created at all, it is controlled by the player who chose to make it
// rather than by the caster (CR 707.10b), and its target is the one
// that player picked.

const (
	ChainOfVaporOracle = "2167ac25-d042-4b48-b770-8b94acc1a965"
	ChainOfSmogOracle  = "ea14c26b-bf2f-48b4-b879-6e63069ded1f"
)

// pushLand puts a land on the battlefield under a player's control.
func PushLand(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Basic Land — Island",
		Owner: owner, Controller: owner,
	})
	return id
}

// copiesOnStack returns the stack items flagged IsCopy.
func CopiesOnStack(g *game.Game) []game.StackItem {
	var out []game.StackItem
	for _, c := range g.Stack.Cards {
		item := g.StackItemForEffect(c.InstanceID)
		if item == nil || !item.IsCopy {
			continue
		}
		out = append(out, *item)
	}
	return out
}
