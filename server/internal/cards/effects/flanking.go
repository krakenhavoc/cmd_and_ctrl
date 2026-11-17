package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flanking.go — Flanking (CR 702.25a): "Whenever this creature becomes
// blocked by a creature without flanking, the blocking creature gets
// -1/-1 until end of turn." Append-only, mechanic-named.
//
// Since #5784 flanking is an ENGINE keyword trigger (game/flanking.go):
// one trigger per flanking token on the creature's layered ability list,
// so a printed, imported or granted flanking triggers (Agility), each
// instance separately (CR 702.25b), and a creature that loses flanking
// (Barbed Foliage) does not. A card that prints flanking declares only
// the token: PrintedKeywords: []string{game.KeywordFlanking}.
//
// The row below is RETIRED. It was the trigger before #5784, and the
// cards that shipped with it keep it, in the same place on their
// Triggered list, because a restore point written before #5784 may hold
// that row's trigger on the stack: the item names the row by key and
// index (ADR 0041 P9, #1497), resolves through it, and a missing row
// would flag the card for the rest of the game (#522). The row never
// triggers now, so the keyword path is the only one that does and no
// instance triggers twice. Never add it to a new card;
// TestOnlyTheLegacyFlankingCardsKeepTheRetiredRow keeps it to the cards
// that had it.

// RetiredFlankingRow is the pre-#5784 flanking row for the card named
// `name`: it keeps its key (the restore identity) and its resolution,
// and never triggers. Keyword marks it as the flanking keyword's, so the
// ability tiles show the keyword's chip and not a second row.
func RetiredFlankingRow(name string) game.TriggeredAbility {
	label := fmt.Sprintf("%s — flanking", name)
	row := On(game.EventBlock, func(game.Event, *game.Card, game.Characteristic, *game.Game) bool {
		return false
	}, label, flankingShrink)
	row.Keyword = game.KeywordFlanking
	return row
}

// flankingShrink gives the blocker the trigger saw -1/-1 until end of
// turn: the retired row's resolution, for a trigger restored from
// before #5784. A blocker that has left the battlefield is gone (CR
// 400.7).
func flankingShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
		return nil
	}
	return BoostUntilEOT{Target: ctx.Trigger().Event.CardID, Power: -1, Toughness: -1}.Apply(ctx)
}
