package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b752_helpers.go — shared vocabulary for the S82 OLD batch 752 cards
// (Tempest, #5628).

// OldB752DurationWhileItIsEnchanted is "for as long as that creature is
// enchanted" (Rootwater Matriarch), pinned to `id` as the object it is now (CR
// 400.7): the effect lasts while at least one Aura is attached to it. False, so
// the card registers nothing, when it is not enchanted as the effect would begin
// (CR 611.2b).
func OldB752DurationWhileItIsEnchanted(ctx *Context, id uuid.UUID) (game.Duration, bool) {
	return ctx.Game.ForAsLongAsPinnedEnchantedDuration(id)
}

// OldB752BasicLandWalkKeywords is the landwalk keywords for the basic land types
// of a permanent as it last existed on the battlefield (CR 603.10a), read from
// the leaves-the-battlefield event of the card an ability's cost sacrificed:
// "landwalk of each of the land types of the sacrificed land" (Excavator). A card
// with no such event is read as it is printed.
func OldB752BasicLandWalkKeywords(g *game.Game, ref game.ObjectRef) []string {
	card, ok := g.LookupCardForEffect(ref.ID)
	if !ok {
		return nil
	}
	var ltb *game.Event
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == game.EventLTB && g.Events[i].CardID == ref.ID {
			ltb = &g.Events[i]
			break
		}
	}
	var out []string
	for _, t := range []struct{ subtype, keyword string }{
		{"Plains", "plainswalk"}, {"Island", "islandwalk"}, {"Swamp", "swampwalk"},
		{"Mountain", "mountainwalk"}, {"Forest", "forestwalk"},
	} {
		has := card.HasSubtype(t.subtype)
		if ltb != nil {
			has = LeftAsSubtype(*ltb, card, t.subtype)
		}
		if has {
			out = append(out, t.keyword)
		}
	}
	return out
}
