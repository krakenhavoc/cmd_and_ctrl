package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// play_from_outside.go — "cast ... cards you own from outside the game"
// and "play a card you own from outside the game" (#3955, ADR 0198's
// 2026-10-16 amendment): the cast half of the seam TakeFromOutside
// (outside_game.go) is the bring-in half of.

// PlayFromOutside asks the owner for the cards, from the same pool
// TakeFromOutside reads (their sideboard, or the card database when the
// deck seated none), and grants a permission over each one in the
// container the pick ends in. The cast itself is an ordinary cast_spell
// with `from_zone: "outside"`, taken afterwards from the sideboard tab.
//
//   - Free is Spawnsire of Ulamog's "without paying their mana costs":
//     a free, any-timing, cast-only, pass-closed cast of each card
//     (game.FreeCastFromOutsidePermission). The cards not cast when the
//     holder passes stay outside the game.
//   - not Free is Wish's "play": until end of turn, at the card's own
//     timing, for its own cost, a land included
//     (game.PlayFromOutsidePermission).
//
// Max 0 is "any number" (game.OutsideAnyNumber); Wish says Max 1.
type PlayFromOutside struct {
	// Player owns the cards and chooses. Zero means the resolving
	// effect's controller.
	Player uuid.UUID
	// Match narrows the pool ("an Eldrazi card"); nil offers every card.
	Match func(game.Card) bool
	// Name labels the permission and heads the question: the card's name.
	Name string
	// Clause is the pick's object, written the way the card is: "any
	// number of Eldrazi cards", "a card".
	Clause string
	// Free is the cast without paying the mana cost.
	Free bool
	// Max bounds the pick; zero is any number.
	Max int
}

func (t PlayFromOutside) Apply(ctx *Context) error {
	player := t.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	source, name, free := ctx.Source(), t.Name, t.Free
	max := t.Max
	if max <= 0 {
		max = game.OutsideAnyNumber
	}
	verb, dest := "play", game.ChoosePlayFromOutside
	if free {
		verb, dest = "cast", game.ChooseCastFreeFromOutside
	}
	question := name + " — choose " + t.Clause + " you own from outside the game to " + verb
	if free {
		question += " without paying their mana costs"
	} else {
		question += " this turn"
	}
	// Frozen scalars only (the StackItem.Effect contract): the closure
	// resolves against whichever *Game it is handed.
	ctx.Game.QueueOutsidePickForEffect(game.OutsidePickPrompt{
		Player:      player,
		Source:      source,
		Question:    question,
		Match:       t.Match,
		Min:         0,
		Max:         max,
		Destination: dest,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			for _, id := range picked {
				if free {
					g.FreeCastFromOutsideForEffect(player, source, id, name)
				} else {
					g.PlayFromOutsideForEffect(player, source, id, name)
				}
			}
			return nil
		},
	})
	return nil
}
