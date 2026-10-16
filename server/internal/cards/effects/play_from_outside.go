package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// play_from_outside.go — "cast ... cards you own from outside the game"
// and "play a card you own from outside the game" (#3955, ADR 0198's
// 2026-10-16 amendment): the cast half of the seam TakeFromOutside
// (outside_game.go) is the bring-in half of.

// PlayFromOutside is Spawnsire of Ulamog's "cast any number of Eldrazi
// cards you own from outside the game without paying their mana costs".
// It asks the owner for the cards, from the same pool TakeFromOutside
// reads (their sideboard, or the card database when the deck seated
// none), and grants a free, any-timing, cast-only, pass-closed cast over
// each picked card in the container the pick ends in
// (game.FreeCastFromOutsidePermission). The cards not cast when the
// holder passes stay outside the game. The cast itself is an ordinary
// cast_spell with `from_zone: "outside"`, taken from the sideboard tab.
//
// The pick is "any number" (game.OutsideAnyNumber), and the same card
// name may be chosen more than once from the database: a player may own
// two copies of a card.
type PlayFromOutside struct {
	// Player owns the cards and chooses. Zero means the resolving
	// effect's controller.
	Player uuid.UUID
	// Match narrows the pool ("an Eldrazi card"); nil offers every card.
	Match func(game.Card) bool
	// Name labels the permission and heads the question: the card's name.
	Name string
	// Clause is the pick's object, written the way the card is: "any
	// number of Eldrazi cards".
	Clause string
}

func (t PlayFromOutside) Apply(ctx *Context) error {
	player := t.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	source, name := ctx.Source(), t.Name
	question := name + " — choose " + t.Clause + " you own from outside the game to cast without paying their mana costs"
	// Frozen scalars only (the StackItem.Effect contract): the closure
	// resolves against whichever *Game it is handed.
	ctx.Game.QueueOutsidePickForEffect(game.OutsidePickPrompt{
		Player:      player,
		Source:      source,
		Question:    question,
		Match:       t.Match,
		Min:         0,
		Max:         game.OutsideAnyNumber,
		Destination: game.ChooseCastFreeFromOutside,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			for _, id := range picked {
				g.FreeCastFromOutsideForEffect(player, source, id, name)
			}
			return nil
		},
	})
	return nil
}

// GrantOutsidePlayRight is Wish's "you may play a card you own from
// outside the game this turn": it names nothing. It stores a standing,
// one-use, until-end-of-turn permission over the owner's container
// (game.OutsidePlayRightPermission), and the card is chosen LATER, at the
// moment the player wants to play it, at whatever timing that card
// allows. In a sideboard pool the owner casts a sideboard card directly;
// in the database pool the `outside_play` action opens the pick, mints the
// named card, and the cast or land play follows at once. Cancelling that
// pick spends nothing: the permission is spent by a card played.
type GrantOutsidePlayRight struct {
	// Player owns the cards. Zero means the resolving effect's controller.
	Player uuid.UUID
	// Name labels the permission: the card's name.
	Name string
}

func (t GrantOutsidePlayRight) Apply(ctx *Context) error {
	player := t.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	ctx.Game.GrantOutsidePlayRightForEffect(player, ctx.Source(), t.Name)
	return nil
}
