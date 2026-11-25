package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b111_helpers.go — shapes Ikoria: Lair of Behemoths batch 111
// (#5512) prints: "the number of times this creature has mutated" and the
// tokens, drains and returns the mutate triggers make.

// S82OldB111TimesMutated is "the number of times this creature has mutated"
// (Huntmaster Liger, Insatiable Hemophage, Auspicious Starrix), read as the
// trigger resolves: each merge adds one component to the permanent, so a
// permanent made of N cards has mutated N-1 times (CR 702.140d, CR 730.2). A
// permanent that has left the battlefield by then is counted from the turn's
// log of merges into it, since its last entry; the merge that triggered the
// ability is always counted, so the answer is at least one.
func S82OldB111TimesMutated(g *game.Game, id uuid.UUID) int {
	if c, ok := g.LookupCardForEffect(id); ok && OnBattlefield(g, id) && c.IsMerged() {
		return max(len(c.MergedFrom)-1, 1)
	}
	n := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.CardID != id {
			continue
		}
		if ev.Kind == game.EventMutate {
			n++
		}
		if ev.Kind == game.EventZoneMove && ev.NewZone == game.ZoneBattlefield {
			break
		}
	}
	return max(n, 1)
}

// S82OldB111XOfThisMutation is the X of a mutate trigger worded "where X is
// the number of times this creature has mutated", for the resolving ability.
func S82OldB111XOfThisMutation(ctx *Context) int {
	return S82OldB111TimesMutated(ctx.Game, ctx.Source())
}

// S82OldB111FirstGraveyardTarget is the first legal target card of the
// resolving trigger, when it is still a creature card in the graveyard.
func S82OldB111FirstGraveyardTarget(ctx *Context) (uuid.UUID, bool) {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return uuid.Nil, false
	}
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return uuid.Nil, false
	}
	return id, true
}
