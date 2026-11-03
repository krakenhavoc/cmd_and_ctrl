package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b12_helpers.go — shared pieces of Secrets of Strixhaven Commander,
// batch 12 (#4834).

// socB12YouCastAnInstantOrSorceryThisTurn is the trigger-time half of an
// intervening "if you've cast an instant or sorcery spell this turn"
// (CR 603.4): Rootha's check as the beginning of combat starts.
func socB12YouCastAnInstantOrSorceryThisTurn(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.CastTallyFor(source.Controller).InstantSorcery > 0
}

// socB12ExileTargetsAndCopyThem exiles the still-legal card targets of
// the resolving trigger through the CR 614 window and hands the cards that
// really reached exile to `then`: Renegade Bull's "exile up to one target
// card from your graveyard and copy it".
func socB12ExileTargetsAndCopyThem(g *game.Game, item *game.StackItem, then game.TailRef) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(then))
}
