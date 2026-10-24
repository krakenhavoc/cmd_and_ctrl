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
//
// The pick is keyed (#4256, registerPlayFromOutside), so a table waiting
// on it is a restore point.
type PlayFromOutside struct {
	// Player owns the cards and chooses. Zero means the resolving
	// effect's controller.
	Player uuid.UUID
	// Pick is the card's registered pick: its filter, its name (which
	// labels the permission and heads the question) and its key.
	// Required.
	Pick outsidePlay
	// Word is the one word of the card's text the filter reads
	// (Spawnsire's "Eldrazi", which a text-changing effect may have
	// changed). Plain data on the prompt.
	Word string
	// Clause is the pick's object, written the way the card is: "any
	// number of Eldrazi cards".
	Clause string
}

// outsidePlay is one card's keyed cast-from-outside pick.
type outsidePlay struct {
	pick game.OutsidePick
	name string
}

// registerPlayFromOutside registers a card's cast-from-outside pick under
// `key`: `match` is the filter (nil offers every card), handed the word
// the prompt carries, and `name` labels each grant. The name is bound to
// the key rather than read off the source when the answer comes in,
// freeHandCastOffer's reason: the source may be gone by then. Call it
// once, from a package-level var.
func registerPlayFromOutside(key, name string, match game.OutsideMatchFunc) outsidePlay {
	return outsidePlay{
		name: name,
		pick: game.RegisterOutsidePick(key, match, func(g *game.Game, r game.CardsChosen) error {
			if !r.Answered {
				return nil
			}
			for _, id := range g.StillOutsideForEffect(r.Chooser, r.Picked) {
				g.FreeCastFromOutsideForEffect(r.Chooser, r.Source, id, name)
			}
			return nil
		}),
	}
}

func (t PlayFromOutside) Apply(ctx *Context) error {
	player := t.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	question := t.Pick.name + " — choose " + t.Clause + " you own from outside the game to cast without paying their mana costs"
	ctx.Game.QueueOutsidePickForEffect(game.OutsidePickPrompt{
		Player:      player,
		Source:      ctx.Source(),
		Question:    question,
		Min:         0,
		Max:         game.OutsideAnyNumber,
		Destination: game.ChooseCastFreeFromOutside,
		Keyed:       t.Pick.pick,
		MatchArg:    t.Word,
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
