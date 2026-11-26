package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b457_helpers.go — shared seams of Commander 2011, batch 457 (#5574).

// S82OldB457AnOpponentControlsLandsAtLeast is the cost condition "an opponent
// controls <n> or more lands" (Avatar of Fury): ONE opponent controls at least
// n lands on their own, in a four-player game the opponents' lands are not
// added together. Lands are counted through the layers, as the cost is
// determined (CR 601.2f).
func S82OldB457AnOpponentControlsLandsAtLeast(n int) CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		return EachOpponent(q.Game, q.Controller, func(opponent uuid.UUID) bool {
			return CountControlled(q.Game, opponent, func(c game.Card) bool { return c.IsLand() }) >= n
		})
	}
}

// S82OldB457TuckThisAndWhatItBlocks is Gomazoa's ability: "Put this creature and
// each creature it's blocking on top of their owners' libraries, then those
// players shuffle their libraries." The creatures are read as the ability
// resolves: Gomazoa itself if it is still on the battlefield, and every attacker
// it is blocking (CR 509.1a) that still is. They go on top as one simultaneous
// move and each owner shuffles once afterwards, Void Stalker's tail.
func S82OldB457TuckThisAndWhatItBlocks(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok && OnBattlefield(g, c.InstanceID) && c.IsCreature() {
		ids = append(ids, c.InstanceID)
		blocked := append([]uuid.UUID{c.BlockingTarget}, c.AlsoBlocking...)
		for _, id := range blocked {
			if id == uuid.Nil {
				continue
			}
			if b, ok := g.LookupCardForEffect(id); ok && OnBattlefield(g, id) && b.IsCreature() {
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{}, ctx.Tail(s82OldB429VoidStalkerShuffle))
}
