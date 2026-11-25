package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b109_helpers.go — shapes Core Set 2021 batch 109 (#5511) prints:
// Chrome Replicator's "same name as one another" count, Turn to Slag's
// "destroy all Equipment attached to that creature" and Transmogrify's
// "shuffles the rest" tail.

// S82OldB109TwoSameNamePermanents is "you control two or more nonland,
// nontoken permanents with the same name as one another" (Chrome
// Replicator): some name, read through the layers, shared by two of your
// nonland nontoken permanents.
func S82OldB109TwoSameNamePermanents(g *game.Game, you uuid.UUID) bool {
	seen := map[string]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you || c.IsToken() || c.HasCardType("land") {
			continue
		}
		seen[c.Effective().Name]++
		if seen[c.Effective().Name] >= 2 {
			return true
		}
	}
	return false
}

// S82OldB109DestroyAttachedEquipment is "Destroy all Equipment attached to
// that creature" (Turn to Slag): every Equipment permanent whose host is
// `host` as the instruction is read.
func S82OldB109DestroyAttachedEquipment(ctx *Context, host uuid.UUID) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.AttachedTo.ID == host && c.AttachedTo.Kind == game.TargetCard && c.HasSubtype("Equipment") {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (DestroyTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB109ShuffleRestTail is "then shuffles the rest into their library"
// (Transmogrify): the revealed cards never left the library, so the rest is
// a shuffle. An on-disk key: never renamed, never reused.
var S82OldB109ShuffleRestTail = PutFromLibraryTail("put-from-library/s82-old-b109-shuffle-the-rest",
	func(ctx *Context, res PutFromLibraryResult) error {
		return ctx.Game.ShuffleLibraryForEffect(res.Player)
	})
