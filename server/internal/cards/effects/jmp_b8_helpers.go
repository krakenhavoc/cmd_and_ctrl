package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b8_helpers.go — bodies for the Jumpstart Azorius deck request's
// eighth batch (#4894). Append-only.

// returnExiledWithAPlusOneCounterBody is the delayed return of "exile
// target creature. At the beginning of the next end step, return that
// card to the battlefield under its owner's control with a +1/+1
// counter on it" (Long Road Home).
var returnExiledWithAPlusOneCounterBody = game.SimpleDelayedBody(
	"flicker/return-exiled-with-a-plus-one-counter", returnExiledCardsWithAPlusOneCounter)

// returnExiledCardsWithAPlusOneCounter returns the listed cards from
// exile together under their owners' control, each entering with a
// +1/+1 counter whatever it is as it enters: unlike Semester's End, the
// clause has no "if it's a creature". The counter is the entry's own
// (ZoneEntryOptions.EntersWithCounters), so an enters trigger sees it
// and a counter replacement applies to it (CR 122.6). A card that left
// exile before the trigger resolved is not followed (CR 603.7c), and a
// token never comes back (CR 111.8).
func returnExiledCardsWithAPlusOneCounter(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			ids = append(ids, t.ID)
		}
	}
	return putCardsFromZoneTogetherWith(NewContext(g, item), game.ZoneExile, ids,
		game.ZoneEntryOptions{EntersWithCounters: map[string]int{game.CounterPlusOne: 1}}, game.TailRef{}, 0)
}
