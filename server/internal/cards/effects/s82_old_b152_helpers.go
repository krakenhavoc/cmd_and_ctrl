package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b152_helpers.go — helpers for older-set batch 152 (Commander 2019,
// #5517). A file of its own, so concurrent batches never collide.

// OldB152CreaturesPutIntoYourGraveyardFromTheBattlefieldThisTurn is "each
// creature put into your graveyard from the battlefield this turn" (Fresh
// Meat): every creature, token or not, that went from the battlefield to
// `you`'s graveyard this turn, judged on its last-known types and counted once
// per trip. A card is put into its owner's graveyard (CR 400.3), so a creature
// of yours an opponent controlled counts and one of theirs you controlled does
// not. A token has ceased to exist (CR 704.5d) and cannot be looked up, so its
// event stands for itself: its last-known types say it was a creature, and the
// player who controlled it as it died is the one whose graveyard it was put
// into (a token's owner is its creator).
func OldB152CreaturesPutIntoYourGraveyardFromTheBattlefieldThisTurn(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok {
			if LeftAsType(ev, c, "creature") && c.Owner == you {
				n++
			}
			continue
		}
		if was, known := ev.WasType("creature"); known && was && ev.Actor == you {
			n++
		}
	}
	return n
}
