package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Approach of the Second Sun — Sorcery {6}{W} (EDHREC rank 968):
//
//	"If this spell was cast from your hand and you've cast another spell
//	 named Approach of the Second Sun this game, you win the game.
//	 Otherwise, put Approach of the Second Sun into its owner's library
//	 seventh from the top and you gain 7 life."
//
// Both conditions are read as it resolves (the 2017-07-14 ruling):
//
//   - "Cast from your hand" is the item's CastFromZone. A copy was not
//     cast at all (CR 707.10), so a copy always takes the "otherwise".
//   - "Another spell named Approach" is read off the game's event log:
//     every EventCast its controller made of a card with that name,
//     less this cast, which is the latest cast of this card. The first
//     one may have been cast from anywhere and need not have resolved,
//     and the same card cast on an earlier turn counts, because a card
//     that changes zones is a new object (the 2017-04-18 rulings).
//
// The win is WinTheGame (ADR 0057). Otherwise the spell tucks itself
// seventh from the top (TuckOptions.Depth, the bottom of a library of
// six or fewer) and its controller gains 7 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4125377-34c0-4b54-bdf8-4e88f5d24565",
		Name:         "Approach of the Second Sun",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if item.CastFromZone == game.ZoneHand &&
				approachCastsThisGame(ctx.Game, ctx.Controller(), ctx.Source()) >= 1 {
				return WinTheGame{}.Apply(ctx)
			}
			return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{ctx.Source()}, game.TuckOptions{Depth: 7}, ctx.ItemTail(approachOfTheSecondSunGain))
		},
	})
}

// approachName is the name the first sentence asks about.
const approachName = "Approach of the Second Sun"

// approachCastsThisGame counts the spells named Approach of the Second
// Sun `player` has cast this game OTHER than this one: the latest
// EventCast of `self` is this spell's own cast and is not counted.
func approachCastsThisGame(g *game.Game, player, self uuid.UUID) int {
	n := 0
	skippedSelf := false
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind != game.EventCast || ev.Actor != player {
			continue
		}
		if ev.CardID == self && !skippedSelf {
			skippedSelf = true
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if ok && game.CardNameMatches(c, approachName) {
			n++
		}
	}
	return n
}

// approachOfTheSecondSunGain is "you gain 7 life" after the tuck (#4253 slice 9).
var approachOfTheSecondSunGain = CardTail("approach-of-the-second-sun/gain-seven", func(ctx *Context, o game.TailOutcome) error {
	g, item, controller := ctx.Game, ctx.Item, ctx.Controller()

	return g.ChangePlayerLifeForEffect(item.SourceCardID, controller, 7)
})
