package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// recruit.go — the recruit keyword action (CR 701.70, #4475):
//
//	701.70a "Recruit" means "Draw a card, then discard a card. If you
//	  discarded a nonland card this way, create a 1/1 white Human Soldier
//	  creature token."
//
// (September 25, 2026 edition.) It is connive's shape (connive.go) with a
// token where connive puts a counter, so it is built the same way: draw,
// then a real player-chosen discard whose continuation is a registered
// key (#4253), so a table waiting on the discard is a restore point with
// nothing captured, and the token is made from what really left the hand.
//
// The recruiting player is the one the effect names: the ability's
// controller for "recruit" and "you recruit" alike (The Queen of Dale
// recruits on an opponent's turn, still for her controller). The draw
// comes first and the discard is answered before the token, so a card
// drawn by the recruit may be the one discarded. An empty hand after the
// draw discards nothing (CR 701.9a) and makes no token; a land discarded
// makes none either.
//
// Usable from any resolution: a triggered ability's Do(Recruit{}), a
// spell's clause, a Saga chapter. Recruit is the last instruction of every
// card that has it today, so nothing runs after the discard but the token.

// Recruit makes Player recruit. A zero Player is the context's controller,
// which is "recruit" on every card printed with it.
type Recruit struct {
	Player uuid.UUID
}

// Apply draws a card and queues the discard whose answer makes the token.
func (r Recruit) Apply(ctx *Context) error {
	player := r.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	if player == uuid.Nil {
		return nil
	}
	var source uuid.UUID
	if ctx.Item != nil {
		source = ctx.Item.SourceCardID
	}
	return recruitFor(ctx.Game, player, source)
}

// recruitFor is the body: draw, then the keyed discard.
func recruitFor(g *game.Game, player, source uuid.UUID) error {
	if err := g.DrawNForEffect(player, 1); err != nil {
		return err
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   player,
		Source:   source,
		N:        1,
		Question: "Recruit: discard a card (a nonland card makes a 1/1 Human Soldier)",
	}, recruitDiscarded, nil)
}

// recruitSoldier is the token CR 701.70a names.
func recruitSoldier() game.Card { return TokenCard("1/1 white Human Soldier") }

// recruitDiscarded makes the Human Soldier for a nonland card discarded.
// An on-disk key: never renamed, never reused.
var recruitDiscarded = game.RegisterDiscardThen("choose-cards/recruit-discarded", func(g *game.Game, r game.DiscardsDone) error {
	discarded := r.Discarded.By(r.Player)
	if len(discarded) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(discarded[0])
	if !ok || card.IsLand() {
		return nil
	}
	return CreateToken{Controller: r.Player, Template: recruitSoldier(), N: 1}.Apply(keyedContext(g, r.Player, r.Source))
})
