package effectstest

import (
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_of_leng_shuffle_test.go — #3541: one discard puts some cards
// on top of the library (Library of Leng) and shuffles another into it
// (Blightsteel Colossus). CR 701.24g: the library is shuffled and THEN
// the top cards are put on top, and CR 401.4 lets their owner order
// exactly those cards. The legs settle one after another, so a later
// shuffle used to scatter the earlier top placements and the batch then
// offered no order at all.

// lengSettleBatch answers the discard batch's replacement prompts as it
// walks: each Leng "may" by `accept` for the card being discarded (the
// first of `order` still in hand), and each CR 616 order with Library of
// Leng first. It stops at the first prompt that is neither, and reports
// how many CR 616 order prompts it answered.
func LengSettleBatch(t *testing.T, g *game.Game, me *game.Player, order []uuid.UUID, accept map[uuid.UUID]bool) int {
	t.Helper()
	orders := 0
	for step := 0; step < 4*len(order)+4; step++ {
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && (pc.Kind == game.PendingChoiceOptionalReplacement || pc.Kind == game.PendingChoiceReplacementOrder) {
				c = pc
				break
			}
		}
		if c == nil {
			return orders
		}
		var cur uuid.UUID
		for _, id := range order {
			if me.Hand.Contains(id) {
				cur = id
				break
			}
		}
		switch c.Kind {
		case game.PendingChoiceReplacementOrder:
			// Library of Leng is a battlefield card's replacement, so
			// its ID is in the catalog range, below every card's own
			// ("reveal it and shuffle it in instead") replacement.
			ids := append([]game.ReplacementEffectID(nil), c.ReplacementEffectIDs...)
			sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
			if err := g.ResolveReplacementOrder(c.ID, me.ID, ids); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}
			orders++
		case game.PendingChoiceOptionalReplacement:
			if err := g.ResolveOptionalReplacement(c.ID, me.ID, accept[cur]); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
		}
	}
	t.Fatal("the discard batch never stopped asking")
	return orders
}
